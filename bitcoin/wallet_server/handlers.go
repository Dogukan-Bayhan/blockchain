package main

import (
	"bitcoin/block"
	"bitcoin/utils"
	"bitcoin/wallet"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
)

// Index serves the wallet HTML page.
func (ws *WalletServer) Index(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		templatePath := path.Join(tempDir, "index.html")
		if _, err := os.Stat(templatePath); err != nil {
			templatePath = path.Join("templates", "index.html")
		}

		t, err := template.ParseFiles(templatePath)
		if err != nil {
			log.Printf("Error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := t.Execute(w, ""); err != nil {
			log.Printf("Error: %v", err)
		}
	default:
		log.Printf("Error: Invalid Http Method")
	}
}

// Wallet creates a new wallet and returns its keys and address.
func (ws *WalletServer) Wallet(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		w.Header().Add("Content-Type", "application/json")
		myWallet := wallet.NewWallet()
		m, err := myWallet.MarshalJSON()
		if err != nil {
			log.Printf("Error: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			if _, writeErr := w.Write(utils.JsonStatus("fail")); writeErr != nil {
				log.Printf("Error: %v", writeErr)
			}
			return
		}
		if _, err := w.Write(m); err != nil {
			log.Printf("Error: %v", err)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Error: Invalid Http method")
	}
}

// CreateTransaction signs a wallet transaction and submits it to the blockchain node.
func (ws *WalletServer) CreateTransaction(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(req.Body)
		var t wallet.TransactionRequest
		err := decoder.Decode(&t)
		if err != nil {
			log.Printf("Error: %v", err)
			w.Write(utils.JsonStatus("fail"))
			return
		}

		if !t.Validate() {
			log.Println("Error: missing field(s)")
			w.Write(utils.JsonStatus("fail"))
			return
		}

		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		if publicKey == nil {
			log.Println("Error: invalid public key")
			w.Write(utils.JsonStatus("fail"))
			return
		}
		privateKey := utils.PrivateKeyFromString(*t.SenderPrivateKey, publicKey)
		if privateKey == nil {
			log.Println("Error: invalid private key")
			w.Write(utils.JsonStatus("fail"))
			return
		}
		value, err := strconv.ParseFloat(*t.Value, 32)
		if err != nil {
			log.Printf("Error: parse value: %v", err)
			w.Write(utils.JsonStatus("fail"))
			return
		}
		value32 := float32(value)

		w.Header().Add("Content-Type", "application/json")

		// The wallet signs the transaction locally before forwarding only the
		// public transaction request to the blockchain server.
		transaction := wallet.NewTransaction(privateKey, publicKey,
			*t.SenderBlockchainAddress, *t.RecipientBlockchainAddress, value32)

		signature := transaction.GenerateSignature()
		if signature == nil {
			log.Println("Error: generate signature")
			w.Write(utils.JsonStatus("fail"))
			return
		}
		signatureStr := signature.String()

		bt := &block.TransactionRequest{
			t.SenderBlockchainAddress,
			t.RecipientBlockchainAddress,
			t.SenderPublicKey,
			&value32, &signatureStr,
		}

		m, err := json.Marshal(bt)
		if err != nil {
			log.Printf("Error: %v", err)
			w.Write(utils.JsonStatus("fail"))
			return
		}
		buf := bytes.NewBuffer(m)

		resp, err := http.Post(ws.Gateway()+"/transactions", "application/json", buf)
		if err != nil {
			log.Printf("Error: %v", err)
			w.Write(utils.JsonStatus("fail"))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 201 {
			if _, err := io.WriteString(w, string(utils.JsonStatus("success"))); err != nil {
				log.Printf("Error: %v", err)
			}
			return
		}
		if _, err := io.WriteString(w, string(utils.JsonStatus("fail"))); err != nil {
			log.Printf("Error: %v", err)
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Error: Invalid Http method")
	}
}

// WalletAmount proxies balance lookup requests to the configured blockchain node.
func (ws *WalletServer) WalletAmount(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		blockchainAddress := req.URL.Query().Get("blockchain_address")
		endpoint := fmt.Sprintf("%s/amount", ws.Gateway())

		client := &http.Client{}
		bcsReq, _ := http.NewRequest("GET", endpoint, nil)
		q := bcsReq.URL.Query()
		q.Add("blockchain_address", blockchainAddress)
		bcsReq.URL.RawQuery = q.Encode()

		bcsResp, err := client.Do(bcsReq)
		if err != nil {
			log.Printf("ERROR: %v", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		defer bcsResp.Body.Close()

		w.Header().Add("Content-Type", "application/json")
		if bcsResp.StatusCode == 200 {
			decoder := json.NewDecoder(bcsResp.Body)
			var bar block.AmountResponse
			err := decoder.Decode(&bar)
			if err != nil {
				log.Printf("ERROR: %v", err)
				io.WriteString(w, string(utils.JsonStatus("fail")))
				return
			}

			m, _ := json.Marshal(struct {
				Message string  `json:"message"`
				Amount  float32 `json:"amount"`
			}{
				Message: "success",
				Amount:  bar.Amount,
			})
			io.WriteString(w, string(m[:]))
		} else {
			io.WriteString(w, string(utils.JsonStatus("fail")))
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Error: Invalid Http method")
	}
}
