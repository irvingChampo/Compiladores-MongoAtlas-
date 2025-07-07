package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mongoapi/config"
)

type ExecuteRequest struct {
	Comando string `json:"comando"`
}

type ExecuteResponse struct {
	Exito   bool   `json:"exito"`
	Mensaje string `json:"mensaje"`
}

func ExecuteHandler(w http.ResponseWriter, r *http.Request) {
	var req ExecuteRequest
	json.NewDecoder(r.Body).Decode(&req)
	cmd := strings.TrimSpace(req.Comando)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := config.MongoClient.Database("BaseChampo")

	// CREATE COLLECTION
	if match := matchCreateCollection(cmd); match != "" {
		err := db.CreateCollection(ctx, match)
		responder(w, err == nil, messageOrOK(err, "Colección creada correctamente"))
		return
	}

	// INSERT ONE
	if col, doc := matchInsertOne(cmd); col != "" {
		var docMap bson.M
		err := bson.UnmarshalExtJSON([]byte(doc), true, &docMap)
		if err != nil {
			responder(w, false, "Error al parsear documento: "+err.Error())
			return
		}
		result, err := db.Collection(col).InsertOne(ctx, docMap)
		if err != nil {
			responder(w, false, "Error al insertar: "+err.Error())
		} else {
			responder(w, true, fmt.Sprintf("Documento insertado con ID: %v", result.InsertedID))
		}
		return
	}

	// FIND ONE
	if col, id := matchFindOne(cmd); col != "" {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			responder(w, false, "ObjectId inválido: "+err.Error())
			return
		}
		var result bson.M
		err = db.Collection(col).FindOne(ctx, bson.M{"_id": objID}).Decode(&result)
		if err != nil {
			responder(w, false, "No se encontró el documento: "+err.Error())
		} else {
			// Convertir a JSON para una mejor presentación
			jsonData, _ := json.MarshalIndent(result, "", "  ")
			responder(w, true, fmt.Sprintf("Documento encontrado:\n%s", string(jsonData)))
		}
		return
	}

	// UPDATE ONE
	if col, id, doc := matchUpdateOne(cmd); col != "" {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			responder(w, false, "ObjectId inválido: "+err.Error())
			return
		}
		
		// Convertir el documento de actualización a BSON
		update, err := parseUpdateDocument(doc)
		if err != nil {
			responder(w, false, "Error al parsear actualización: "+err.Error())
			return
		}
		
		result, err := db.Collection(col).UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			responder(w, false, "Error al actualizar: "+err.Error())
		} else {
			responder(w, true, fmt.Sprintf("Documento actualizado. Documentos modificados: %d", result.ModifiedCount))
		}
		return
	}

	// DELETE ONE
	if col, id := matchDeleteOne(cmd); col != "" {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			responder(w, false, "ObjectId inválido: "+err.Error())
			return
		}
		result, err := db.Collection(col).DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			responder(w, false, "Error al eliminar: "+err.Error())
		} else {
			responder(w, true, fmt.Sprintf("Documento eliminado. Documentos eliminados: %d", result.DeletedCount))
		}
		return
	}

	// DROP COLLECTION
	if col := matchDropCollection(cmd); col != "" {
		err := db.Collection(col).Drop(ctx)
		responder(w, err == nil, messageOrOK(err, "Colección eliminada"))
		return
	}

	// DROP DATABASE
	if strings.Contains(cmd, "db.dropDatabase") {
		err := db.Drop(ctx)
		responder(w, err == nil, messageOrOK(err, "Base de datos eliminada"))
		return
	}

	// LIST COLLECTIONS
	if strings.Contains(cmd, "getCollectionNames") {
		names, err := db.ListCollectionNames(ctx, bson.D{})
		if err != nil {
			responder(w, false, "Error al listar colecciones")
		} else {
			responder(w, true, "Colecciones: "+strings.Join(names, ", "))
		}
		return
	}

	responder(w, false, "Comando no reconocido")
}

// ------------ REGEX MATCHERS (MEJORADOS) ------------

func matchCreateCollection(cmd string) string {
	re := regexp.MustCompile(`db\.createCollection\(\s*["'](\w+)["']\s*\)`)
	matches := re.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func matchInsertOne(cmd string) (string, string) {
	re := regexp.MustCompile(`db\.(\w+)\.insertOne\(\s*(\{.*\})\s*\)`)
	matches := re.FindStringSubmatch(cmd)
	if len(matches) > 2 {
		return matches[1], matches[2]
	}
	return "", ""
}

func matchFindOne(cmd string) (string, string) {
	// Más flexible para diferentes formatos de espacios
	re := regexp.MustCompile(`db\.(\w+)\.findOne\(\s*\{\s*_id\s*:\s*ObjectId\(\s*["']([a-fA-F0-9]{24})["']\s*\)\s*\}\s*\)`)
	matches := re.FindStringSubmatch(cmd)
	if len(matches) > 2 {
		return matches[1], matches[2]
	}
	return "", ""
}

func matchUpdateOne(cmd string) (string, string, string) {
	// Patrón que captura correctamente documentos anidados con $set
	re := regexp.MustCompile(`db\.(\w+)\.updateOne\(\s*\{\s*_id\s*:\s*ObjectId\(\s*["']([a-fA-F0-9]{24})["']\s*\)\s*\}\s*,\s*(\{(?:[^{}]|\{[^{}]*\})*\})\s*\)`)
	matches := re.FindStringSubmatch(cmd)
	if len(matches) > 3 {
		return matches[1], matches[2], matches[3]
	}
	return "", "", ""
}

func matchDeleteOne(cmd string) (string, string) {
	re := regexp.MustCompile(`db\.(\w+)\.deleteOne\(\s*\{\s*_id\s*:\s*ObjectId\(\s*["']([a-fA-F0-9]{24})["']\s*\)\s*\}\s*\)`)
	matches := re.FindStringSubmatch(cmd)
	if len(matches) > 2 {
		return matches[1], matches[2]
	}
	return "", ""
}

func matchDropCollection(cmd string) string {
	re := regexp.MustCompile(`db\.(\w+)\.drop\(\s*\)`)
	matches := re.FindStringSubmatch(cmd)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// ------------ UTILIDADES ------------

func parseUpdateDocument(doc string) (bson.M, error) {
	// Primero intentar parsearlo como Extended JSON
	var update bson.M
	err := bson.UnmarshalExtJSON([]byte(doc), true, &update)
	if err == nil {
		return update, nil
	}
	
	// Si falla, intentar parsearlo como JSON estándar y luego procesarlo
	var rawDoc map[string]interface{}
	err = json.Unmarshal([]byte(doc), &rawDoc)
	if err != nil {
		return nil, fmt.Errorf("formato de documento inválido: %v", err)
	}
	
	// Convertir a bson.M
	update = make(bson.M)
	for key, value := range rawDoc {
		update[key] = value
	}
	
	// Si no contiene operadores de actualización, envolver en $set
	if !containsUpdateOperators(update) {
		update = bson.M{"$set": update}
	}
	
	return update, nil
}

func containsUpdateOperators(doc bson.M) bool {
	for key := range doc {
		if strings.HasPrefix(key, "$") {
			return true
		}
	}
	return false
}

// ------------ RESPUESTA ------------

func responder(w http.ResponseWriter, ok bool, msg string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ExecuteResponse{
		Exito:   ok,
		Mensaje: msg,
	})
}

func messageOrOK(err error, msg string) string {
	if err != nil {
		return err.Error()
	}
	return msg
}