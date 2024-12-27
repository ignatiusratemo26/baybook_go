package controllers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"baybook_go/data"
	"baybook_go/models"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MpesaCredentials struct {
	AccessToken string
	ExpiresIn   string
}

type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
}

func generateAccessToken() (string, error) {
	consumerKey := os.Getenv("MPESA_CONSUMER_KEY")
	consumerSecret := os.Getenv("MPESA_CONSUMER_SECRET")

	// Validate credentials exist
	if consumerKey == "" || consumerSecret == "" {
		return "", fmt.Errorf("missing Mpesa credentials")
	}

	// Create auth string and encode to base64
	auth := base64.StdEncoding.EncodeToString([]byte(consumerKey + ":" + consumerSecret))

	url := "https://sandbox.safaricom.co.ke/oauth/v1/generate?grant_type=client_credentials"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Cache-Control", "no-cache")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// Log the response for debugging
	log.Printf("Mpesa API Response: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var mpesaCredentials struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   string `json:"expires_in"`
	}

	err = json.Unmarshal(body, &mpesaCredentials)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if mpesaCredentials.AccessToken == "" {
		return "", fmt.Errorf("no access token in response")
	}

	return mpesaCredentials.AccessToken, nil
}

func InitiateMpesaPayment(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers if needed
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse request body
	var paymentRequest struct {
		BookingID   string `json:"booking_id"`
		PhoneNumber string `json:"phone_number"`

		Amount float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&paymentRequest); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Log the request for debugging
	log.Printf("Payment request received: %+v", paymentRequest)

	// Convert booking ID from string to ObjectID
	bookingID, err := primitive.ObjectIDFromHex(paymentRequest.BookingID)
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	// Create initial payment record
	payment := models.MpesaPayment{
		ID:            primitive.NewObjectID(),
		BookingID:     bookingID,
		Amount:        paymentRequest.Amount,
		PaymentMethod: "MPESA",
		Status:        models.PaymentStatusPending,
		PhoneNumber:   paymentRequest.PhoneNumber,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Get MongoDB collection
	paymentsCollection := data.GetMongoClient().Database("baybookDB").Collection("payments")

	// Insert payment record
	_, err = paymentsCollection.InsertOne(context.Background(), payment)
	if err != nil {
		http.Error(w, "Failed to create payment record", http.StatusInternalServerError)
		return
	}

	// Get access token
	accessToken, err := generateAccessToken()
	if err != nil {
		log.Printf("Error generating access token: %v", err)
		http.Error(w, fmt.Sprintf("Failed to generate access token: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare STK push request
	timestamp := time.Now().Format("20060102150405")
	businessShortCode := os.Getenv("MPESA_SHORTCODE")
	passkey := os.Getenv("MPESA_PASSKEY")

	// Generate password
	password := base64.StdEncoding.EncodeToString([]byte(businessShortCode + passkey + timestamp))

	stkPushURL := "https://sandbox.safaricom.co.ke/mpesa/stkpush/v1/processrequest"

	requestBody := map[string]interface{}{
		"BusinessShortCode": businessShortCode,
		"Password":          password,
		"Timestamp":         timestamp,
		"TransactionType":   "CustomerPayBillOnline",
		"Amount":            payment.Amount,
		"PartyA":            payment.PhoneNumber,
		"PartyB":            businessShortCode,
		"PhoneNumber":       payment.PhoneNumber,
		"CallBackURL":       os.Getenv("MPESA_CALLBACK_URL"),
		"AccountReference":  "BayBook Booking",
		"TransactionDesc":   "Payment for accommodation booking",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		http.Error(w, "Failed to prepare request", http.StatusInternalServerError)
		return
	}

	// Make STK push request
	req, err := http.NewRequest("POST", stkPushURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var stkPushResponse STKPushResponse
	err = json.NewDecoder(resp.Body).Decode(&stkPushResponse)
	if err != nil {
		http.Error(w, "Failed to decode response", http.StatusInternalServerError)
		return
	}

	// Update payment record with Mpesa response
	update := bson.M{
		"$set": bson.M{
			"merchant_request_id": stkPushResponse.MerchantRequestID,
			"checkout_request_id": stkPushResponse.CheckoutRequestID,
			"updated_at":          time.Now(),
		},
	}

	_, err = paymentsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": payment.ID},
		update,
	)

	if err != nil {
		log.Printf("Failed to update payment record: %v", err)
	}

	// Return response to client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stkPushResponse)
}

func MpesaCallback(w http.ResponseWriter, r *http.Request) {
	var callbackData struct {
		Body struct {
			StkCallback struct {
				MerchantRequestID string `json:"MerchantRequestID"`
				CheckoutRequestID string `json:"CheckoutRequestID"`
				ResultCode        int    `json:"ResultCode"`
				ResultDesc        string `json:"ResultDesc"`
				CallbackMetadata  struct {
					Item []struct {
						Name  string      `json:"Name"`
						Value interface{} `json:"Value"`
					} `json:"Item"`
				} `json:"CallbackMetadata"`
			} `json:"stkCallback"`
		} `json:"Body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&callbackData); err != nil {
		http.Error(w, "Invalid callback data", http.StatusBadRequest)
		return
	}

	// Get MongoDB collection
	paymentsCollection := data.GetMongoClient().Database("baybookDB").Collection("payments")

	// Find the payment record
	var payment models.MpesaPayment
	err := paymentsCollection.FindOne(
		context.Background(),
		bson.M{"merchant_request_id": callbackData.Body.StkCallback.MerchantRequestID},
	).Decode(&payment)

	if err != nil {
		log.Printf("Payment record not found: %v", err)
		w.WriteHeader(http.StatusOK) // Still return 200 to Mpesa
		return
	}

	// Extract Mpesa receipt number and other metadata
	var mpesaReceiptNo string
	for _, item := range callbackData.Body.StkCallback.CallbackMetadata.Item {
		if item.Name == "MpesaReceiptNumber" {
			mpesaReceiptNo = item.Value.(string)
			break
		}
	}

	// Update payment status
	status := models.PaymentStatusFailed
	if callbackData.Body.StkCallback.ResultCode == 0 {
		status = models.PaymentStatusCompleted
	}

	update := bson.M{
		"$set": bson.M{
			"status":           status,
			"mpesa_receipt_no": mpesaReceiptNo,
			"updated_at":       time.Now(),
		},
	}

	_, err = paymentsCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": payment.ID},
		update,
	)

	if err != nil {
		log.Printf("Failed to update payment status: %v", err)
	}

	// If payment was successful, update booking status
	if status == models.PaymentStatusCompleted {
		bookingsCollection := data.GetMongoClient().Database("baybookDB").Collection("bookings")
		_, err = bookingsCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": payment.BookingID},
			bson.M{"$set": bson.M{"payment_status": "PAID"}},
		)
		if err != nil {
			log.Printf("Failed to update booking status: %v", err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

// Add this helper function to get payment status
func GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paymentID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid payment ID", http.StatusBadRequest)
		return
	}

	paymentsCollection := data.GetMongoClient().Database("baybookDB").Collection("payments")
	var payment models.MpesaPayment
	err = paymentsCollection.FindOne(context.Background(), bson.M{"_id": paymentID}).Decode(&payment)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Payment not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}
