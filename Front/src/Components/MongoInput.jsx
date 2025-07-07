import React, { useState } from 'react'
import '../Styles/MongoInput.css'

const MongoInput = ({ label, ejemplo, index, onAnalizar, disabled }) => {
  const [texto, setTexto] = useState("")

  const handleClick = () => {
    if (texto.trim() !== "") {
      onAnalizar(index, texto)
    }
  }

  return (
    <div className="input-box">
      <label>{label}</label>
      <small>Ejemplo: {ejemplo}</small>
      <textarea
        placeholder="Escribe aquí el comando..."
        value={texto}
        onChange={(e) => setTexto(e.target.value)}
        disabled={disabled}
      />
      <button onClick={handleClick} disabled={disabled || texto.trim() === ""}>
        Analizar y Ejecutar
      </button>
    </div>
  )
}

export default MongoInput
