package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

// url 3ady bta3 el connect aleh
const URL = "mongodb://localhost:27017"

func main() {
	// abd2 a3ml el hat3mlo da bel condtion da be duration kza hy3ml timeout b3d ad eh
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)

	//
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URL))
	if err != nil {
		panic(err)
	}

	defer func() {
		// lw el client disconnected bel db error variable false /true
		err := client.Disconnect(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()
	// ba5od el hagat mn el mongdab store in result set
	resultSet, _ := client.ListDatabaseNames(ctx, bson.D{})
	// print all database like for each
	for _, db := range resultSet {
		fmt.Println(db)
	}
	database := client.Database("Clinic")        // Replace with your database name
	collection := database.Collection("Doctors") // Replace with your collection name

	// Doctor set slots ----------------------------------------------------------------------------------
	// doctor enter his data name,date,time
	var DocName string
	fmt.Print("enter doc name")
	fmt.Scan(&DocName)
	var date string
	fmt.Print("enter date")
	fmt.Scan(&date)
	var timee string
	fmt.Print("enter time")
	fmt.Scan(&timee)
	//*****************************************************************************************************************
	// check on data and time format
	dateLayout := "01/02/2006"
	timeLayout := "3:04PM"
	_, err = time.Parse(dateLayout, date)
	if err != nil {
		fmt.Println("Invalid date format:", err)
		return
	}
	_, err = time.Parse(timeLayout, timee)
	if err != nil {
		fmt.Println("Invalid time format:", err)
		return
	}
	//*****************************************************************************************************************
	// store info of doctor into mongodb and check info
	document := bson.D{
		{"Doctor", DocName},
		{"date", date},
		{"time", timee},
		{"IsAvailable", true},
	}
	_, err = collection.InsertOne(context.Background(), document)
	if err != nil {
		fmt.Println("Failed to insert document into MongoDB:", err)
		return
	}
	fmt.Println("Data inserted into MongoDB.")
	//*****************************************************************************************************************
	//Patients select doctor, view his available slots, then patient chooses a slot.-----------------------------------
	// patient enter doc name and his name
	var PatientName string
	fmt.Println("Enter Patient name: ")
	fmt.Scan(&PatientName)

	var DoctorName string
	fmt.Println("Enter doctor name: ")
	fmt.Scan(&DoctorName)
	//*****************************************************************************************************************
	// check on available slot of this doctore and return all available slot
	availableSlots, err := getAvailableSlots(ctx, collection, DoctorName)
	if err != nil {
		fmt.Println("Failed to get available slots:", err)
		return
	}

	fmt.Println("Available slots for", DoctorName, ":")
	for _, slot := range availableSlots {
		fmt.Println(slot)
	}
	//***************************************************************************************************************
	// choose slot
	var selectedSlotTime string
	var selectedSlotdate string
	fmt.Print("Choose a Time&date: ")
	fmt.Scan(&selectedSlotTime, &selectedSlotdate)
	_, err = time.Parse(timeLayout, selectedSlotTime)
	_, err = time.Parse(dateLayout, date)
	if err != nil {
		fmt.Println("Invalid date format:", err)
		return
	}
	if err != nil {
		fmt.Println("Invalid time format:", err)
		return
	}
	//****************************************************************************************************************
	// Update the selected slot's availability to false
	filter := bson.M{
		"Doctor":      DoctorName,
		"date":        selectedSlotdate,
		"time":        selectedSlotTime,
		"IsAvailable": true,
	}
	update := bson.M{
		"$set": bson.M{"IsAvailable": false},
	}
	updateResult, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		fmt.Println("Failed to update slot availability:", err)
		return
	}

	if updateResult.ModifiedCount == 0 {
		fmt.Println("Slot is already taken or does not exist.")
		return
	}
	//****************************************************************************************************************
	// register slot into appointment
	appointmentsCollection := database.Collection("Appointments")
	appointment := bson.D{
		{"Doctor", DoctorName},
		{"Time", selectedSlotTime},
		{"Date", selectedSlotdate},
		{"PatientName", PatientName},
	}
	_, err = appointmentsCollection.InsertOne(context.Background(), appointment)
	if err != nil {
		fmt.Println("Failed to schedule the appointment:", err)
		return
	}
	fmt.Println("Appointment scheduled successfully.")
	//*****************************************************************************************************************
	//Patient can update his appointment by change the doctor or the slot.---------------------------------------------
	var patientName string
	fmt.Print("Enter the patient name : ")
	fmt.Scan(&patientName)
	// Query the MongoDB collection to find the patient's appointment
	Newfilter := bson.M{
		"PatientName": patientName,
	}
	var Oldappointment struct {
		ID          primitive.ObjectID `bson:"_id,omitempty"`
		Doctor      string             `bson:"Doctor"`
		Date        string             `bson:"Date"`
		Time        string             `bson:"Time"`
		PatientName string             `bson:"PatientName"`
	}
	err = appointmentsCollection.FindOne(ctx, Newfilter).Decode(&Oldappointment)
	if err != nil {
		fmt.Println("Appointment not found:", err)
		return
	}
	fmt.Printf("Your current appointment: Doctor: %s, date: %s, time: %s", Oldappointment.Doctor, Oldappointment.Date, Oldappointment.Time)
	//**************************************************************************************************************************************************
	// Ask the patient what they want to update
	fmt.Println("What would you like to update?")
	fmt.Println("1. Doctor")
	fmt.Println("2. Slot")
	fmt.Print("Enter your choice (1 or 2): ")
	var choice int
	fmt.Scan(&choice)
	//******************************************************************************************************************
	//1.change doctor
	if choice == 1 {
		var newDoctorName string
		fmt.Print("Enter the new doctor's name: ")
		fmt.Scan(&newDoctorName)
		//**************************************************************************************************************
		// reserve new slot for new doctor and return all available slot
		availableSlots, err := getAvailableSlots(ctx, collection, newDoctorName)
		if err != nil {
			fmt.Println("Failed to get available slots:", err)
			return
		}

		fmt.Println("Available slots for", newDoctorName, ":")
		for _, slot := range availableSlots {
			fmt.Println(slot)
		}
		fmt.Print("Choose a date&time: ")
		fmt.Scan(&selectedSlotTime, &selectedSlotdate)
		_, err = time.Parse(timeLayout, selectedSlotTime)
		_, err = time.Parse(dateLayout, selectedSlotdate)
		if err != nil {
			fmt.Println("Invalid date format:", err)
			return
		}
		if err != nil {
			fmt.Println("Invalid time format:", err)
			return
		}
		//**************************************************************************************************************
		// update state of new doc from is available =true to false
		filter = bson.M{
			"Doctor":      newDoctorName,
			"date":        selectedSlotdate,
			"time":        selectedSlotTime,
			"IsAvailable": true,
		}
		update = bson.M{
			"$set": bson.M{"IsAvailable": false},
		}
		updateResult, err = collection.UpdateOne(ctx, filter, update)
		//fmt.Println("Filter:", filter)
		//fmt.Println("Update:", update)
		if err != nil {
			fmt.Println("Failed to update slot availability:", err)
			return
		}
		//************************************************************************************************************
		// return old slot of old to true. make is available to true again
		filterNew := bson.M{
			"Doctor":      Oldappointment.Doctor,
			"date":        Oldappointment.Date,
			"time":        Oldappointment.Time,
			"IsAvailable": false,
		}
		fmt.Println("Filter:", filterNew)
		update = bson.M{
			"$set": bson.M{"IsAvailable": true},
		}
		updateResult, err = collection.UpdateOne(ctx, filterNew, update)
		fmt.Println("Update:", update)
		if err != nil {
			fmt.Println("Failed to update slot availability:", err)
			return
		}
		// update appoitments in appoitments collection
		update = bson.M{
			"$set": bson.M{"Doctor": newDoctorName, "Date": selectedSlotdate, "Time": selectedSlotTime},
		}
		updateResult, err = appointmentsCollection.UpdateOne(ctx, Newfilter, update)
		if err != nil {
			fmt.Println("Failed to update appointment:", err)
			return
		}

		if updateResult.ModifiedCount == 0 {
			fmt.Println("No appointment updated.")
		} else {
			fmt.Println("Appointment updated successfully.")
		}
		//*************************************************************************************************************
	} else if choice == 2 {
		//*************************************************************************************************************
		// find available slot of old doctor then choose it
		availableSlots, err = getAvailableSlots(ctx, collection, Oldappointment.Doctor)
		if err != nil {
			fmt.Println("Failed to get available slots:", err)
			return
		}

		fmt.Println("Available slots for", Oldappointment.Doctor, ":")
		for _, slot := range availableSlots {
			fmt.Println(slot)
		}
		fmt.Print("Choose a date&time: ")
		fmt.Scan(&selectedSlotTime, &selectedSlotdate)
		_, err = time.Parse(timeLayout, selectedSlotTime)
		_, err = time.Parse(dateLayout, selectedSlotdate)
		if err != nil {
			fmt.Println("Invalid date format:", err)
			return
		}
		if err != nil {
			fmt.Println("Invalid time format:", err)
			return
		}
		//************************************************************************************************************
		// Check if the new slot is available
		if !isSlotAvailable(selectedSlotdate, selectedSlotTime, availableSlots) {
			fmt.Println("Slot is not available.")
			return
		}
		//************************************************************************************************************
		// update new slot in appointment
		update = bson.M{
			"$set": bson.M{"Date": selectedSlotdate, "Time": selectedSlotTime},
		}
		updateResult, err = appointmentsCollection.UpdateOne(ctx, Newfilter, update)
		if err != nil {
			fmt.Println("Failed to update appointment:", err)
			return
		}
		if updateResult.ModifiedCount == 0 {
			fmt.Println("No appointment updated.")
		} else {
			fmt.Println("Appointment updated successfully.")
		}
		//************************************************************************************************************
		// return old slot of old to true. make is available to true again
		filterNew := bson.M{
			"Doctor":      Oldappointment.Doctor,
			"date":        Oldappointment.Date,
			"time":        Oldappointment.Time,
			"IsAvailable": false,
		}
		fmt.Println("Filter:", filterNew)
		update = bson.M{
			"$set": bson.M{"IsAvailable": true},
		}
		updateResult, err = collection.UpdateOne(ctx, filterNew, update)
		fmt.Println("Update:", update)
		if err != nil {
			fmt.Println("Failed to update slot availability:", err)
			return
		}
		//*************************************************************************************************************
		// update state of new doc from is available =true to false
		filter = bson.M{
			"Doctor":      Oldappointment.Doctor,
			"date":        selectedSlotdate,
			"time":        selectedSlotTime,
			"IsAvailable": true,
		}
		update = bson.M{
			"$set": bson.M{"IsAvailable": false},
		}
		updateResult, err = collection.UpdateOne(ctx, filter, update)
		//fmt.Println("Filter:", filter)
		//fmt.Println("Update:", update)
		if err != nil {
			fmt.Println("Failed to update slot availability:", err)
			return
		}
		//**************************************************************************************************************

	} else {
		fmt.Println("Invalid choice. Please enter 1 or 2.")
	}

}
func getAvailableSlots(ctx context.Context, collection *mongo.Collection, doctorName string) ([]string, error) {
	filter := bson.M{
		"Doctor":      doctorName,
		"IsAvailable": true,
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var availableSlots []string
	for cursor.Next(ctx) {
		var slot struct {
			Time  string `bson:"time"`
			Datee string `bson:"date"`
		}
		if err := cursor.Decode(&slot); err != nil {
			return nil, err
		}
		availableSlots = append(availableSlots, slot.Time, slot.Datee)
	}

	return availableSlots, nil
}

// Function to check if a slot is available
func isSlotAvailable(date string, time string, availableSlots []string) bool {
	for _, s := range availableSlots {
		if s == time || s == date {
			return true
		}
	}
	return false
}
