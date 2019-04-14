package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

/**
* objeto json
 */
type Enunciado struct {
	NumeroCasas         int    `json:"numero_casas"`
	Token               string `json:"token"`
	Cifrado             string `json:"cifrado"`
	Decifrado           string `json:"decifrado"`
	ResumoCriptografico string `json:"resumo_criptografico"`
}

type MultipartForm struct {
	Data *multipart.FileHeader `json:"data"`
}

func main() {
	var bytesjson []byte

	// 1. recuperando o json da API
	response, err := http.Get("https://api.codenation.dev/v1/challenge/dev-ps/generate-data?token=79208e3125419816d61f18fd7c755b6a900dfb83")
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		bytesjson, _ = ioutil.ReadAll(response.Body)
	}

	var enunciado Enunciado
	const letraA, letraZ, qtdLetras = 'a', 'z', 26

	// 2. descriptografando
	json.Unmarshal(bytesjson, &enunciado)
	enunciado.Decifrado = strings.ToLower(enunciado.Decifrado)
	decipher := func(letraCifrada rune) rune {
		var letraDecifrada = letraCifrada
		if letraCifrada >= 'a' && letraCifrada <= 'z' {
			calculo := letraCifrada - 'a' - rune(enunciado.NumeroCasas)
			if calculo <= 0 {
				calculo = calculo + qtdLetras
			}
			letraDecifrada = rune((calculo % qtdLetras) + 'a')
		}
		return letraDecifrada
	}
	enunciado.Decifrado = strings.Map(decipher, enunciado.Cifrado)

	// 3. gerando o hash
	hash := sha1.New()
	hash.Write([]byte(enunciado.Cifrado))
	enunciado.ResumoCriptografico = hex.EncodeToString(hash.Sum(nil))

	// 4. codificando como json novamente
	json, err := json.Marshal(enunciado)

	// 5. postando o resultado na API
	err = ioutil.WriteFile("answer.json", json, 0644)
	file, err := os.Open("answer.json")
	if err != nil {
		fmt.Printf("Failed with error %s\n", err)
	}
	defer file.Close()

	// fi, err := file.Stat()
	if err != nil {
		fmt.Printf("Failed with error %s\n", err)
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("answer", `answer`)
	if err != nil {
		fmt.Printf("Failed with error %s\n", err)
	}
	_, err = io.Copy(part, file)
	err = writer.Close()

	uri := "https://api.codenation.dev/v1/challenge/dev-ps/submit-solution?token=79208e3125419816d61f18fd7c755b6a900dfb83"
	request, err := http.NewRequest("POST", uri, body)
	ct := writer.FormDataContentType()
	// ct := "multipart/form-data; boundary=cb9458e2e144760087abd"
	request.Header.Add("Content-Type", ct)

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	} else {
		var bodyContent []byte
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Header)
		resp.Body.Read(bodyContent)
		resp.Body.Close()
		fmt.Println(bodyContent)
	}
	// response, err = http.Post(, "multipart/form-data", bytes.NewBuffer(json))

}

//
