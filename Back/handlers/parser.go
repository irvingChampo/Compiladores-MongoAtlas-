package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"mongoapi/config" // Importar el paquete config para acceder al cliente MongoDB
	"go.mongodb.org/mongo-driver/mongo"
)

type AnálisisResultado struct {
	Lexico    []Token   `json:"lexico"`
	Sintaxis  []string  `json:"sintaxis"`
	Semantica []string  `json:"semantica"`
	Valido    bool      `json:"valido"`
}

type AnalizarRequest struct {
	Comando string `json:"comando"`
}

type Token struct {
	Tipo   string `json:"tipo"`
	Lexema string `json:"lexema"`
}

var funcionesMongo = []string{
	"createCollection", "insertOne", "findOne", "updateOne",
	"deleteOne", "drop", "dropDatabase", "getCollectionNames",
}

func AnalizarHandler(w http.ResponseWriter, r *http.Request) {
	var req AnalizarRequest
	json.NewDecoder(r.Body).Decode(&req)

	comando := strings.TrimSpace(req.Comando)
	tokens := analizarLexico(comando)
	erroresSintaxis := analizarSintaxis(comando)
	erroresSemantica := analizarSemantica(comando)

	valido := len(erroresSintaxis) == 0 && len(erroresSemantica) == 0

	res := AnálisisResultado{
		Lexico:    tokens,
		Sintaxis:  erroresSintaxis,
		Semantica: erroresSemantica,
		Valido:    valido,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func analizarLexico(comando string) []Token {
	var tokens []Token

	// Mejorar la expresión regular para capturar mejor ObjectId
	re := regexp.MustCompile(`ObjectId\("[a-fA-F0-9]{24}"\)|[\w\.]+|\(|\)|\{|\}|\[|\]|,|:|"[^"]*"|\$\w+`)
	matches := re.FindAllString(comando, -1)

	for _, m := range matches {
		switch {
		case m == "db":
			tokens = append(tokens, Token{"Keyword", m})
		case strings.HasPrefix(m, "ObjectId"):
			tokens = append(tokens, Token{"ObjectId", m})
		case strings.HasPrefix(m, "$"):
			tokens = append(tokens, Token{"Operador MongoDB", m})
		case inSlice(funcionesMongo, m):
			tokens = append(tokens, Token{"Función MongoDB", m})
		case m == ".":
			tokens = append(tokens, Token{"Operador", m})
		case m == "{" || m == "}":
			tokens = append(tokens, Token{"Llave", m})
		case m == "(" || m == ")":
			tokens = append(tokens, Token{"Paréntesis", m})
		case m == "," || m == ":":
			tokens = append(tokens, Token{"Separador", m})
		case strings.HasPrefix(m, "\"") && strings.HasSuffix(m, "\""):
			tokens = append(tokens, Token{"Cadena", m})
		case isNumber(m):
			tokens = append(tokens, Token{"Número", m})
		default:
			tokens = append(tokens, Token{"Identificador", m})
		}
	}
	return tokens
}

func analizarSintaxis(comando string) []string {
	var errores []string

	if !strings.HasPrefix(comando, "db.") {
		errores = append(errores, "El comando debe comenzar con 'db.'")
	}

	if strings.Count(comando, "(") != strings.Count(comando, ")") {
		errores = append(errores, "Paréntesis desbalanceados")
	}

	if strings.Count(comando, "{") != strings.Count(comando, "}") {
		errores = append(errores, "Llaves desbalanceadas")
	}

	ok := false
	for _, fn := range funcionesMongo {
		if strings.Contains(comando, fn) {
			ok = true
			break
		}
	}
	if !ok {
		errores = append(errores, "No se detectó una función válida de MongoDB")
	}

	return errores
}

func analizarSemantica(comando string) []string {
	var errores []string

	// Obtener el cliente MongoDB y la base de datos
	db := config.Client.Database("BaseChampo")

	// Extraer el nombre de la colección (si aplica)
	parts := strings.Split(comando, ".")
	if len(parts) < 3 {
		if !strings.Contains(comando, "createCollection") && !strings.Contains(comando, "getCollectionNames") && !strings.Contains(comando, "dropDatabase") {
			errores = append(errores, "Se esperaba una colección en el comando")
			return errores
		}
	}
	coleccion := ""
	if len(parts) >= 2 {
		coleccion = strings.Split(parts[1], "(")[0]
	}

	// Verificar existencia de la colección para comandos que lo requieren
	if strings.Contains(comando, "insertOne") || strings.Contains(comando, "findOne") || 
	   strings.Contains(comando, "updateOne") || strings.Contains(comando, "deleteOne") || 
	   strings.Contains(comando, "drop") {
		if coleccion != "" && !coleccionExiste(db, coleccion) {
			errores = append(errores, "La colección '"+coleccion+"' no existe")
		}
	}

	// Validaciones semánticas existentes
	if strings.Contains(comando, "findOne") {
		if strings.Contains(comando, "_id:") && !strings.Contains(comando, "ObjectId(") {
			errores = append(errores, "Para buscar por _id se requiere ObjectId")
		}
		if strings.Contains(comando, "ObjectId(") && !validarFormatoObjectId(comando) {
			errores = append(errores, "Formato de ObjectId inválido (debe tener 24 caracteres hexadecimales)")
		}
	}

	if strings.Contains(comando, "updateOne") {
		if strings.Contains(comando, "_id:") && !strings.Contains(comando, "ObjectId(") {
			errores = append(errores, "Para actualizar por _id se requiere ObjectId")
		}
		if !strings.Contains(comando, "$set") && !strings.Contains(comando, "$unset") && 
		   !strings.Contains(comando, "$inc") && !strings.Contains(comando, "$push") {
			if strings.Count(comando, "{") < 2 {
				errores = append(errores, "updateOne requiere un documento de actualización (ej: {$set: {...}})")
			}
		}
		if strings.Contains(comando, "ObjectId(") && !validarFormatoObjectId(comando) {
			errores = append(errores, "Formato de ObjectId inválido")
		}
	}

	if strings.Contains(comando, "deleteOne") {
		if strings.Contains(comando, "_id:") && !strings.Contains(comando, "ObjectId(") {
			errores = append(errores, "Para eliminar por _id se requiere ObjectId")
		}
		if strings.Contains(comando, "ObjectId(") && !validarFormatoObjectId(comando) {
			errores = append(errores, "Formato de ObjectId inválido")
		}
	}

	if strings.Contains(comando, "insertOne") {
		if !strings.Contains(comando, "{") || !strings.Contains(comando, "}") {
			errores = append(errores, "insertOne requiere un documento para insertar")
		}
	}

	return errores
}

func coleccionExiste(db *mongo.Database, coleccion string) bool {
	colecciones, err := db.ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		return false
	}
	for _, c := range colecciones {
		if c == coleccion {
			return true
		}
	}
	return false
}

func validarFormatoObjectId(comando string) bool {
	re := regexp.MustCompile(`ObjectId\("([a-fA-F0-9]{24})"\)`)
	matches := re.FindStringSubmatch(comando)
	return len(matches) > 1 && len(matches[1]) == 24
}

func isNumber(s string) bool {
	re := regexp.MustCompile(`^\d+(\.\d+)?$`)
	return re.MatchString(s)
}

func inSlice(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}