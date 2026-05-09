package http

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AdminClientsHandler struct {
	db         *sql.DB
	httpClient *http.Client
}

func NewAdminClientsHandler(db *sql.DB) *AdminClientsHandler {
	return &AdminClientsHandler{
		db: db,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (h *AdminClientsHandler) BuscarCliente(w http.ResponseWriter, r *http.Request) {
	documento := strings.TrimSpace(r.URL.Query().Get("documento"))
	if documento == "" {
		writeAPIError(w, http.StatusBadRequest, "No se encontró el documento asociado")
		return
	}

	if cliente, found, err := h.findClienteInDB(documento); err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	} else if found {
		direccion := strings.TrimSpace(cliente.Direccion)
		if direccion == "" {
			direccion = "-"
		}
		writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
			"ok": true,
			"cliente": map[string]string{
				"cliente":   strings.ToUpper(cliente.Denominacion),
				"direccion": strings.ToUpper(direccion),
				"correo":    strings.ToLower(cliente.Correo),
			},
		})
		return
	}

	tipo := 0
	switch len(documento) {
	case 8:
		tipo = 1
	case 11:
		tipo = 2
	default:
		writeAPIError(w, http.StatusBadRequest, "El documento debe tener 8 o 11 dígitos")
		return
	}

	result, err := h.consultarDocumento(tipo, documento)
	if err != nil {
		writeAPIError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeHandlerJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"cliente": result,
	})
}

type clienteDBRecord struct {
	Denominacion string
	Direccion    string
	Correo       string
}

func (h *AdminClientsHandler) findClienteInDB(documento string) (clienteDBRecord, bool, error) {
	if h.db == nil {
		return clienteDBRecord{}, false, nil
	}

	var c clienteDBRecord
	err := h.db.QueryRow(`SELECT denominacion, direccion, correo FROM clientes WHERE documento = ? LIMIT 1`, documento).Scan(&c.Denominacion, &c.Direccion, &c.Correo)
	if err == nil {
		return c, true, nil
	}
	if err == sql.ErrNoRows {
		return clienteDBRecord{}, false, nil
	}
	if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") || strings.Contains(strings.ToLower(err.Error()), "does not exist") {
		return clienteDBRecord{}, false, nil
	}
	return clienteDBRecord{}, false, fmt.Errorf("error al consultar cliente local")
}

func (h *AdminClientsHandler) consultarDocumento(tipo int, documento string) (map[string]string, error) {
	if tipo != 1 && tipo != 2 {
		return nil, fmt.Errorf("El tipo debe ser 1 o 2")
	}
	if strings.TrimSpace(documento) == "" {
		return nil, fmt.Errorf("El documento no puede estar vacío")
	}

	file := "dni.php"
	if tipo == 2 {
		file = "ruc.php"
	}

	url := fmt.Sprintf("http://clientapi.sistemausqay.com/%s?documento=%s", file, documento)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("no se pudo construir consulta externa")
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("no se pudo consultar el documento")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("respuesta inválida del servicio externo")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("respuesta inválida del servicio externo")
	}

	if tipo == 1 {
		if _, ok := payload["dni"]; !ok {
			return nil, fmt.Errorf("Ocurrió un error al consultar el documento")
		}
		apellidos, _ := payload["apellidos"].(string)
		nombres, _ := payload["nombres"].(string)
		direccion, _ := payload["direccion"].(string)
		direccion = strings.TrimSpace(direccion)
		if direccion == "" {
			direccion = "-"
		}
		return map[string]string{
			"cliente":   strings.TrimSpace(strings.TrimSpace(apellidos) + " " + strings.TrimSpace(nombres)),
			"direccion": direccion,
			"correo":    "",
		}, nil
	}

	if _, ok := payload["ruc"]; !ok {
		return nil, fmt.Errorf("Ocurrió un error al consultar el documento")
	}
	razonSocial, _ := payload["razon_social"].(string)
	direccion, _ := payload["direccion"].(string)
	direccion = strings.TrimSpace(direccion)
	if direccion == "" {
		direccion = "-"
	}
	return map[string]string{
		"cliente":   razonSocial,
		"direccion": direccion,
		"correo":    "",
	}, nil
}
