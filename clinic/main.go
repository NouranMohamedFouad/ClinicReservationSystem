package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"time"
)

type User struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Username string             `json:"username,omitempty" bson:"username,omitempty"`
	Password string             `json:"password,omitempty" bson:"password,omitempty"`
	UserType string             `json:"userType,omitempty" bson:"userType,omitempty"`
}
type Patients struct {
	ID   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name string             `json:"name,omitempty" bson:"name,omitempty"`
}

type Doctors struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"`
	Date        string             `json:"date,omitempty" bson:"date,omitempty"`
	Time        string             `json:"time,omitempty" bson:"time,omitempty"`
	IsAvailable bool               `json:"isAvailable,omitempty" bson:"isAvailable,omitempty"`
}
type Appointments struct {
	ID           primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	PatientName  string             `json:"patientName,omitempty" bson:"patientName,omitempty"`
	DoctorName   string             `json:"doctorName,omitempty" bson:"doctorName,omitempty"`
	SelectedDate string             `json:"date,omitempty" bson:"date,omitempty"`
	SelectedTime string             `json:"time,omitempty" bson:"time,omitempty"`
}
type UpdateDoctors struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	NewName string             `json:"newName,omitempty" bson:"newName,omitempty"`
	Name    string             `json:"name,omitempty" bson:"name,omitempty"`
	Date    string             `json:"date,omitempty" bson:"date,omitempty"`
	Time    string             `json:"time,omitempty" bson:"time,omitempty"`
	OldDate string             `json:"oldDate,omitempty" bson:"oldDate,omitempty"`
	OldTime string             `json:"oldTime,omitempty" bson:"oldTime,omitempty"`
}
type UpdateSlot struct {
	Name    string `json:"doctorName,omitempty" bson:"doctorName,omitempty"`
	Date    string `json:"date,omitempty" bson:"date,omitempty"`
	Time    string `json:"time,omitempty" bson:"time,omitempty"`
	OldDate string `json:"oldDate,omitempty" bson:"oldDate,omitempty"`
	OldTime string `json:"oldTime,omitempty" bson:"oldTime,omitempty"`
}

var client *mongo.Client

//////////////////////////////////////////////////////////////////////////////////////////////////////

// sign up
func SignUPEndPoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var user User
	json.NewDecoder(request.Body).Decode(&user)
	collection := client.Database("Clinic").Collection("Users")
	//ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	//result, _ := collection.InsertOne(ctx, user)
	//json.NewEncoder(response).Encode(result)

	email := user.Username
	password := user.Password
	filter := bson.M{"username": email, "password": password}
	//fmt.Println(filter)
	var results bson.M
	err := collection.FindOne(context.Background(), filter).Decode(&results)
	if err != nil {
		collection := client.Database("Clinic").Collection("Users")
		ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
		result, _ := collection.InsertOne(ctx, user)
		json.NewEncoder(response).Encode(result)
		collectionPatient := client.Database("Clinic").Collection("Patients")
		ctx2, _ := context.WithTimeout(context.Background(), 100*time.Second)
		result2, _ := collectionPatient.InsertOne(ctx2, bson.M{"name": email})
		json.NewEncoder(response).Encode(result2)

	} else if err == nil {
		response.WriteHeader(http.StatusBadRequest)
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////

// sign in
func SignIN(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Decode JSON body
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(`{"error": "Invalid JSON"}`))
		return
	}
	email := input.Username
	password := input.Password
	collection := client.Database("Clinic").Collection("Users")
	filter := bson.M{"username": email, "password": password}
	//fmt.Println(filter)
	var result bson.M
	err := collection.FindOne(context.Background(), filter).Decode(&result)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.WriteHeader(http.StatusBadRequest)
			response.Write([]byte("Invalid credentials"))
		} else {
			log.Printf("error:%v", err)
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{"message": "` + err.Error() + `"}`))
		}
	} else if result["userType"] == "doctor" {
		response.WriteHeader(http.StatusOK)
	} else if result["userType"] == "patient" {

		response.WriteHeader(http.StatusCreated)

	}

}

//////////////////////////////////////////////////////////////////////////////////////////////////////

// set schedule for the doctor
func SetDoctorSchudule(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var Doc Doctors
	json.NewDecoder(request.Body).Decode(&Doc)
	collection := client.Database("Clinic").Collection("Doctors")
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	Doc.IsAvailable = true
	result, _ := collection.InsertOne(ctx, Doc)
	json.NewEncoder(response).Encode(result)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////

// cancel the patient's appointment
func CancelReservation(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var Appointment Appointments
	json.NewDecoder(request.Body).Decode(&Appointment)
	collection := client.Database("Clinic").Collection("Appointments")
	DocCollection := client.Database("Clinic").Collection("Doctors")
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	filter := bson.M{
		"name":        Appointment.DoctorName,
		"date":        Appointment.SelectedDate,
		"time":        Appointment.SelectedTime,
		"isAvailable": false,
	}
	update := bson.M{
		"$set": bson.M{"isAvailable": true},
	}
	updateResult, updateErr := DocCollection.UpdateOne(ctx, filter, update)
	//updateResult, updateErr := DocCollection.UpdateOne(ctx,
	//bson.M{"name": Appointment.DoctorName, "date": Appointment.SelectedDate, "time": Appointment.SelectedTime, "isavailable": false},
	//bson.D{{"$set", bson.M{"isavailable": true}}})
	//fmt.Println("Filter: %+v", bson.M{"name": Appointment.DoctorName, "date": Appointment.SelectedDate, "time": Appointment.SelectedTime})
	if updateErr != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "error": "Failed to update doctor status" }`))
		return
	}
	if updateResult.ModifiedCount == 0 {
		response.WriteHeader(http.StatusNotFound)
		response.Write([]byte(`{ "error": "Doctor not found or status not updated" }`))
		return
	}
	result, err := collection.DeleteOne(ctx, Appointment)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "error": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(result)
}

// ////////////////////////////////////////////////////////////////////////////////////////////////////
func GetAllReservation(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var matchingAppointments []Appointments
	//var Appointment Appointments
	//json.NewDecoder(request.Body).Decode(&Appointment)
	collection := client.Database("Clinic").Collection("Appointments")
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	cursor, err := collection.Find(ctx, bson.M{ /*"patientName": Appointment.PatientName*/ })
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var appointment Appointments
		cursor.Decode(&appointment)
		matchingAppointments = append(matchingAppointments, appointment)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(matchingAppointments)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////

// get all patient's appointments
func GetAllDrSlots(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var matchingSlots []Doctors
	//var Appointment Appointments
	//json.NewDecoder(request.Body).Decode(&Appointment)
	collection := client.Database("Clinic").Collection("Doctors")
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var doc Doctors
		cursor.Decode(&doc)
		matchingSlots = append(matchingSlots, doc)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(matchingSlots)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////

// get available slots
func GetAllSlots(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	//var req ReserveAppointmentRequest
	var doc []Doctors

	collection := client.Database("Clinic").Collection("Doctors")
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	cursor, err := collection.Find(ctx, bson.M{"isAvailable": true})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var doctor Doctors
		cursor.Decode(&doctor)
		doc = append(doc, doctor)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(doc)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////

// create appointment
func ReserveAppointment(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")

	var appointment Appointments
	err := json.NewDecoder(request.Body).Decode(&appointment)
	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(`{ "error": "Invalid JSON" }`))
		return
	}
	// Validate the selected date and time
	if appointment.SelectedDate == "" || appointment.SelectedTime == "" {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(`{ "error": "Selected date and time are required" }`))
		return
	}
	// Store the appointment in the "Appointments" collection
	collection := client.Database("Clinic").Collection("Appointments")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.InsertOne(ctx, appointment)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "error": "` + err.Error() + `" }`))
		return
	}

	response.WriteHeader(http.StatusOK)
	response.Write([]byte(fmt.Sprintf(`{ "message": "Appointment created with ID %s" }`, result.InsertedID)))

	///////////////////////////////////
	filter := bson.M{
		"name":        appointment.DoctorName,
		"date":        appointment.SelectedDate,
		"time":        appointment.SelectedTime,
		"isAvailable": true,
	}
	update := bson.M{
		"$set": bson.M{"isAvailable": false},
	}
	DocCollection := client.Database("Clinic").Collection("Doctors")
	ctx1, _ := context.WithTimeout(context.Background(), 100*time.Second)

	updateResult, err := DocCollection.UpdateOne(ctx1, filter, update)
	if updateResult.ModifiedCount == 0 {
		response.WriteHeader(http.StatusNotFound)
		response.Write([]byte(`{ "error": "appointments not updated" }`))
		return
	}
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "error": "Failed to update slot availability" }`))
		return
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////

// update Appointment
func UpdateReservationDoctor(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var Appointment UpdateDoctors
	json.NewDecoder(request.Body).Decode(&Appointment)
	collection := client.Database("Clinic").Collection("Appointments")
	DocCollection := client.Database("Clinic").Collection("Doctors")
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	filter := bson.M{
		"name":        Appointment.NewName,
		"date":        Appointment.Date,
		"time":        Appointment.Time,
		"isAvailable": true,
	}
	update := bson.M{
		"$set": bson.M{"isAvailable": false},
	}
	updateResult, err := DocCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "error": "Failed to update slot availability" }`))
		return
	}
	// return old slot to true
	filterOldApp := bson.M{
		"name":        Appointment.Name,
		"date":        Appointment.OldDate,
		"time":        Appointment.OldTime,
		"isAvailable": false,
	}
	update = bson.M{
		"$set": bson.M{"isAvailable": true},
	}
	updateResult, err = DocCollection.UpdateOne(ctx, filterOldApp, update)
	fmt.Println("Update:", update)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "error": "Failed to update slot availability" }`))
		return
	}
	// update appointments in appointments collection
	update = bson.M{
		"$set": bson.M{"doctorName": Appointment.NewName, "date": Appointment.Date, "time": Appointment.Time},
	}
	updateResult, err = collection.UpdateOne(ctx, bson.M{"doctorName": Appointment.Name, "date": Appointment.OldDate, "time": Appointment.OldTime}, update)
	fmt.Println(updateResult)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "error": "Failed Appointments" }`))
		return
	}
	if updateResult.ModifiedCount == 0 {
		response.WriteHeader(http.StatusNotFound)
		response.Write([]byte(`{ "error": "appointments not updated" }`))
		return
	} else {
		response.WriteHeader(http.StatusOK)
		response.Write([]byte(fmt.Sprintf(`{ "message": "Appointment Updated Succesfully" }`)))
		return

	}

}

///////////////////////////////////////////////////////////////////////////////////////////////////////

// update the solt for the same doctor
func UpdateReservationSlot(response http.ResponseWriter, request *http.Request) {

	response.Header().Add("content-type", "application/json")
	var Appointment UpdateSlot
	json.NewDecoder(request.Body).Decode(&Appointment)
	collection := client.Database("Clinic").Collection("Appointments")
	DocCollection := client.Database("Clinic").Collection("Doctors")
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)

	// return old slot of old to true. make is available to true again
	filterNew := bson.M{
		"name":        Appointment.Name,
		"date":        Appointment.OldDate,
		"time":        Appointment.OldTime,
		"isAvailable": false,
	}
	fmt.Println(filterNew)
	update := bson.M{
		"$set": bson.M{"isAvailable": true},
	}
	fmt.Println(update)
	updateResult, err := DocCollection.UpdateOne(ctx, filterNew, update)
	fmt.Println(updateResult.ModifiedCount)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "error": "Failed to update slot availability" }`))
		return
	}
	// update state of new doc from is available =true to false
	filter := bson.M{
		"name":        Appointment.Name,
		"date":        Appointment.Date,
		"time":        Appointment.Time,
		"isAvailable": true,
	}
	fmt.Println(filter)
	update = bson.M{
		"$set": bson.M{"isAvailable": false},
	}
	fmt.Println(update)
	updateResult, err = DocCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "error": "Failed to update slot availability" }`))
		return
	}
	fmt.Println(updateResult.ModifiedCount)

	// update new slot in appointment
	update = bson.M{
		"$set": bson.M{"date": Appointment.Date, "time": Appointment.Time},
	}
	fmt.Println(update)
	updateResult, err = collection.UpdateOne(ctx, bson.M{"doctorName": Appointment.Name, "date": Appointment.OldDate, "time": Appointment.OldTime}, update)
	fmt.Println(Appointment.Name, Appointment.OldDate, Appointment.OldTime)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "error": "Failed to update appointment" }`))
		return
	}
	if updateResult.ModifiedCount == 0 {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(`{ "error": "slot not updated" }`))
		return
	} else {
		response.WriteHeader(http.StatusOK)
		response.Write([]byte(fmt.Sprintf(`{ "message": "Appointment Updated Succesfully" }`)))
	}

}

// ////////////////////////////////////////////////
func main() {
	fmt.Println("starting the app")

	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/SignUP", SignUPEndPoint).Methods("POST")
	router.HandleFunc("/SignIN", SignIN).Methods("POST")
	router.HandleFunc("/Doctor/SetSchudule", SetDoctorSchudule).Methods("POST")
	router.HandleFunc("/Doctor/AllSlots", GetAllDrSlots).Methods("GET")

	router.HandleFunc("/Patient/CancelReservation", CancelReservation).Methods("POST")
	router.HandleFunc("/Patient/AllReservation", GetAllReservation).Methods("GET")
	router.HandleFunc("/Patient/Getslot", GetAllSlots).Methods("GET")
	router.HandleFunc("/Patient/ReserveAppointment", ReserveAppointment).Methods("POST")
	router.HandleFunc("/Patient/UpdateReservation/Doctor", UpdateReservationDoctor).Methods("POST")
	router.HandleFunc("/Patient/UpdateReservation/Slot", UpdateReservationSlot).Methods("POST")

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"http://localhost:3000"}) // Replace with the actual origin of your frontend
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	// Use the CORS middleware
	corsHandler := handlers.CORS(originsOk, headersOk, methodsOk)(router)
	// Start the server with the CORS middleware enabled
	//http.ListenAndServe(":3000", corsHandler)
	http.ListenAndServe(":12345", corsHandler)
}
