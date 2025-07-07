import React, { useState } from 'react';
import MongoInput from '../Components/MongoInput';
import '../Styles/Home.css';
import Swal from 'sweetalert2';

// Comandos reales de MongoDB
const acciones = [
  {
    id: 0,
    label: "Crear colección",
    ejemplo: 'db.createCollection("usuarios")'
  },
  {
    id: 1,
    label: "Insertar un documento",
    ejemplo: 'db.usuarios.insertOne({ "nombre": "Champo", "edad": 22 })'
  },
  {
    id: 2,
    label: "Leer documento",
    ejemplo: 'db.usuarios.findOne({ _id: ObjectId("60f6e36b6f1d2c39a8e35a78") })'
  },
  {
    id: 3,
    label: "Eliminar documento",
    ejemplo: 'db.usuarios.deleteOne({ _id: ObjectId("60f6e36b6f1d2c39a8e35a78") })'
  },
  {
    id: 4,
    label: "Listar colecciones",
    ejemplo: 'db.getCollectionNames()'
  },
  {
    id: 5,
    label: "Eliminar colección",
    ejemplo: 'db.usuarios.drop()'
  },
];

const Home = () => {
  const [resultados, setResultados] = useState({
    lexico: [],
    sintaxis: [],
    semantica: []
  });

  const [bloqueados, setBloqueados] = useState(
    Array(acciones.length).fill(true).map((v, i) => i === 0 ? false : true)
  );

  // Usar variable de entorno para la URL del backend
  const API_URL = import.meta.env.VITE_API_URL || 'https://mongo-api-60hj.onrender.com';

  const manejarAnalisis = async (index, comando) => {
    try {
      const res = await fetch(`${API_URL}/api/analizar`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ comando }),
      });

      if (!res.ok) {
        Swal.fire({
          icon: 'error',
          title: 'Error en el análisis',
          text: `Error del servidor: ${res.statusText}`,
          confirmButtonColor: '#0078d4',
        });
        return;
      }

      const data = await res.json();

      setResultados({
        lexico: data.lexico ?? [],
        sintaxis: data.sintaxis ?? [],
        semantica: data.semantica ?? [],
      });

      if (data.valido) {
        if (index + 1 < bloqueados.length) {
          const nuevos = [...bloqueados];
          nuevos[index + 1] = false;
          setBloqueados(nuevos);
        }

        const exec = await fetch(`${API_URL}/api/ejecutar`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ comando }),
        });

        if (!exec.ok) {
          Swal.fire({
            icon: 'error',
            title: 'Error en la ejecución',
            text: `Error del servidor: ${exec.statusText}`,
            confirmButtonColor: '#0078d4',
          });
          return;
        }

        const ejecucion = await exec.json();
        Swal.fire({
          icon: ejecucion.exito ? 'success' : 'error',
          title: ejecucion.exito ? 'Éxito' : 'Error',
          text: ejecucion.mensaje,
          confirmButtonColor: '#0078d4',
        });
      } else {
        Swal.fire({
          icon: 'error',
          title: 'Comando inválido',
          html: `
            <strong>Errores sintácticos:</strong><br>${data.sintaxis.join('<br>') || 'Ninguno'}<br>
            <strong>Errores semánticos:</strong><br>${data.semantica.join('<br>') || 'Ninguno'}
          `,
          confirmButtonColor: '#0078d4',
        });
      }
    } catch (error) {
      Swal.fire({
        icon: 'error',
        title: 'Error de conexión',
        text: `Error al conectar con el servidor: ${error.message}`,
        confirmButtonColor: '#0078d4',
      });
    }
  };

  return (
    <div className="main-container">
      <header className="header">
        <h1>Mongo Atlas</h1>
      </header>
      <div className="contenedor">
        <div className="lado-izquierdo">
          {acciones.map((accion, idx) => (
            <MongoInput
              key={accion.id}
              index={idx}
              label={accion.label}
              ejemplo={accion.ejemplo}
              onAnalizar={manejarAnalisis}
              disabled={bloqueados[idx]}
            />
          ))}
        </div>

        <div className="lado-derecho">
          <div className="seccion-analisis">
            <h3>Análisis Léxico</h3>
            {(resultados.lexico && resultados.lexico.length > 0) ? (
              <table className="lexico-table">
                <thead>
                  <tr>
                    <th>Tipo</th>
                    <th>Lexema</th>
                  </tr>
                </thead>
                <tbody>
                  {resultados.lexico.map((t, i) => (
                    <tr key={i}>
                      <td>{t.tipo}</td>
                      <td>{t.lexema}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <p className="no-data">No hay datos léxicos para mostrar</p>
            )}
          </div>

          <div className="seccion-analisis">
            <h3>Análisis Sintáctico</h3>
            {(resultados.sintaxis && resultados.sintaxis.length > 0) ? (
              resultados.sintaxis.map((s, i) => (
                <p key={i} className="sintaxis-error">{s}</p>
              ))
            ) : (
              <p className="no-error">No hay errores sintácticos</p>
            )}
          </div>

          <div className="seccion-analisis">
            <h3>Análisis Semántico</h3>
            {(resultados.semantica && resultados.semantica.length > 0) ? (
              resultados.semantica.map((s, i) => (
                <p key={i} className="semantica-error">{s}</p>
              ))
            ) : (
              <p className="no-error">No hay errores semánticos</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Home;