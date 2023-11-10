package main

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

const URL = "mongodb://localhost:27017"

// ///////////////////////////////////////////////////////////////////////////////////////

type User struct {
	Username string
	Password string
	UserType string
}

// ///////////////////////////////////////////////////////////////////////////////////////
func signIn(name string, password string, ctx context.Context) (string, error) {

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URL))
	if err != nil {
		return "", err
	}

	defer func() {
		err := client.Disconnect(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	//////////////////////////////////////////////////////////////////////////////////////
	//to print only the date , time ,and doctor name

	usersCollection := client.Database("Clinic").Collection("Users")
	var user User
	err = usersCollection.FindOne(ctx, bson.M{"username": name}).Decode(&user)
	if err != nil {
		return "user not found", err
	}
	if password == user.Password {
		return user.UserType, nil
	}
	return "", errors.New("Invalid password")
}

// ///////////////////////////////////////////////////////////////////////////////////////
func signUp(name string, password string, userTyp string, ctx context.Context) error {

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URL))
	if err != nil {
		panic(err)
	}

	defer func() {
		err := client.Disconnect(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	//////////////////////////////////////////////////////////////////////////////////////
	//to print only the date , time ,and doctor name

	usersCollection := client.Database("Clinic").Collection("Users")
	existingUser := usersCollection.FindOne(ctx, bson.M{"username": name})
	if existingUser.Err() == nil {
		return errors.New("Username already exists")
	}

	newUser := User{
		Username: name,
		Password: password,
		UserType: userTyp,
	}

	// Insert the new user document into the collection
	_, err = usersCollection.InsertOne(ctx, newUser)
	if err != nil {
		return err
	}

	return nil
}

// ///////////////////////////////////////////////////////////////////////////////////////
func setSchedule(DocName string, date string, timee string, number string, ctx context.Context) error {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URL))
	if err != nil {
		return err
	}

	defer func() {
		err := client.Disconnect(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	////////////////////////////////////////////////////
	doctorsCollection := client.Database("Clinic").Collection("Doctors")

	document := bson.D{
		{"name", DocName},
		{"date", date},
		{"time", timee},
		{"contactNumber", number},
		{"isAvailable", true},
	}
	_, err = doctorsCollection.InsertOne(context.Background(), document)
	if err != nil {
		return err
	}

	return nil
}

// ///////////////////////////////////////////////////////////////////////////////////////
func reserveAppointment(DoctorName string, selectedSlotdate string, selectedSlotTime string, PatientName string, ctx context.Context) error {

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URL))
	if err != nil {
		panic(err)
	}

	defer func() {
		err := client.Disconnect(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()
	//part 1 is to make the availability to false
	doctorsCollection := client.Database("Clinic").Collection("Doctors")

	filter := bson.M{
		"name":        DoctorName,
		"date":        selectedSlotdate,
		"time":        selectedSlotTime,
		"isAvailable": "true",
	}
	update := bson.M{
		"$set": bson.M{"isAvailable": "false"},
	}
	updateResult, err := doctorsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		fmt.Println("Failed to update slot availability:", err)
		return err
	}

	if updateResult.ModifiedCount == 0 {
		fmt.Println("Slot is already taken or does not exist.")
		return err
	}
	//******************************************************************************
	// part 2 is to register slot into appointment
	appointmentsCollection := client.Database("Clinic").Collection("Appointments")
	appointment := bson.D{
		{"doctorName", DoctorName},
		{"time", selectedSlotTime},
		{"date", selectedSlotdate},
		{"patientName", PatientName},
	}
	_, err = appointmentsCollection.InsertOne(context.Background(), appointment)
	if err != nil {
		fmt.Println("Failed to schedule the appointment:", err)
		return err
	}
	//////////////////////////////////////////////////////////////////////////////////////
	return nil
}

// ///////////////////////////////////////////////////////////////////////////////////////
//fun update appointment

// ///////////////////////////////////////////////////////////////////////////////////////
func viewAllReservations(patientName string, ctx context.Context) ([]bson.M, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URL))
	if err != nil {
		panic(err)
	}

	defer func() {
		err := client.Disconnect(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	//////////////////////////////////////////////////////////////////////////////////////
	//to print only the date , time ,and doctor name

	appointmentsCollections := client.Database("Clinic").Collection("Appointments")
	projection := bson.M{"doctorName": 1, "date": 1, "time": 1, "_id": 0}

	AppointmentCursor, err := appointmentsCollections.Find(ctx, bson.M{"patientName": patientName}, options.Find().SetProjection(projection))
	if err != nil {
		log.Fatal(err)
	}

	defer AppointmentCursor.Close(ctx)

	var reservations []bson.M
	if err = AppointmentCursor.All(ctx, &reservations); err != nil {
		return nil, err
	}

	return reservations, nil
}

// ///////////////////////////////////////////////////////////////////////////////////////
func cancelAppointment(patientName string, doctorName string, date string, time string, ctx context.Context) error {

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URL))
	if err != nil {
		panic(err)
	}

	defer func() {
		err := client.Disconnect(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()
	//////////////////////////////////////////
	doctorsCollections := client.Database("Clinic").Collection("Doctors")

	appointmentsCollections := client.Database("Clinic").Collection("Appointments")

	deleted, err := appointmentsCollections.DeleteOne(ctx, bson.D{{"patientName", patientName}, {"doctorName", doctorName}, {"time", time}, {"date", date}})
	if err != nil {
		return nil
	}

	doctorsCollections.UpdateOne(
		ctx,
		bson.M{"name": doctorName, "date": date, "time": time},
		bson.D{
			{"$set", bson.M{"isAvailaible": "true"}},
		})

	fmt.Printf("%v Reservations Cancelled Successfully", deleted.DeletedCount)
	fmt.Println(" ")

	return nil
}

// ///////////////////////////////helper function/////////////////////////////////////////
func getAvailableSlots(ctx context.Context, collection *mongo.Collection, doctorName string) ([]string, error) {
	filter := bson.M{
		"name":        doctorName,
		"isAvailable": "true",
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

// ////////////////////////////////helper function/////////////////////////////////////////

func isSlotAvailable(date string, time string, availableSlots []string) bool {
	for _, s := range availableSlots {
		if s == time || s == date {
			return true
		}
	}
	return false
}

// ////////////////////////////////////////////////////////////////////////////////////////
func main() {

	//connecting to database
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)

	defer cancel()

	var choice int
	fmt.Println("Choose an option:")
	fmt.Println("1 - Sign IN")
	fmt.Println("2 - Sign UP")
	fmt.Println("3 - Doctor Set Schedule.")
	fmt.Println("4 - Patient Select Doctor,Reserve an Appointment")
	fmt.Println("5 - Patient Update His Appointment")
	fmt.Println("6 - patient Cancel Appointment")
	fmt.Println("7 - patient View his Reservations")
	fmt.Scan(&choice)

	// check on data and time format
	dateLayout := "01/02/2006"
	timeLayout := "3:04PM"

	switch choice {

	case 1:
		//sign in
		userType, err := signIn("Nouran Mohamed", "54321", ctx)
		if err != nil {
			fmt.Println("Error signing in:", err)
		} else {
			fmt.Println("Signed in as:", userType)
		}

	case 2:
		//sign up
		err := signUp("Nouran Mohamed", "54321", "patient", ctx)
		if err != nil {
			fmt.Println("Error in signing up,This Username Already Exists")
		} else {
			fmt.Println("Signed-Up Successfully")
		}

	case 3:
		// Doctor set slots ----------------------------------------------------------------------------------
		// doctor enter his data name,date,time
		var DocName string
		fmt.Print("enter doc name ")
		fmt.Scan(&DocName)
		var number string
		fmt.Print("enter date ")
		var date string
		fmt.Print("enter date ")
		fmt.Scan(&date)
		var timee string
		fmt.Print("enter time ")
		fmt.Scan(&timee)

		_, err := time.Parse(dateLayout, date)
		if err != nil {
			fmt.Println("Invalid date format:", err)
			return
		}
		_, err = time.Parse(timeLayout, timee)
		if err != nil {
			fmt.Println("Invalid time format:", err)
			return
		}

		errorr := setSchedule(DocName, date, timee, number, ctx)
		if errorr != nil {
			fmt.Println("Failed to insert document into Database:", err)
		} else {
			fmt.Println("Data inserted Successfully.")

		}

	case 4:
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(URL))
		if err != nil {
			panic(err)
		}

		defer func() {
			err := client.Disconnect(ctx)
			if err != nil {
				log.Fatal(err)
			}
		}()

		var PatientName string
		fmt.Println("Enter Patient name: ")
		fmt.Scan(&PatientName)

		var DoctorName string
		fmt.Println("Enter doctor name: ")
		fmt.Scan(&DoctorName)

		doctorsCollection := client.Database("Clinic").Collection("Doctors")
		// check on available slot of this doctor and return all available slot
		availableSlots, err := getAvailableSlots(ctx, doctorsCollection, DoctorName)
		if err != nil {
			fmt.Println("Failed to get available slots:", err)
			return
		}

		fmt.Println("Available slots for", DoctorName, ":")
		for _, slot := range availableSlots {
			fmt.Println(slot)
		}
		//*****************************************************
		// choose slot
		var selectedSlotTime string
		var selectedSlotdate string
		fmt.Print("Choose a Time&date: ")
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
		errorr := reserveAppointment(DoctorName, selectedSlotdate, selectedSlotTime, PatientName, ctx)
		if errorr != nil {
			fmt.Println("Failed to reserve document the Appointment:", err)
		} else {
			fmt.Println("Appointment reserved Successfully.")

		}

		///////////////////////////////////////

	case 5:

	case 6:
		// Call the function to cancel appointment
		err := cancelAppointment("Janna Fattouh", "Dr.Amira", "2023-10-10", "10:00 AM", ctx)
		if err != nil {
			fmt.Println("Error cancelling appointment:", err)
		} else {
			fmt.Println("This Slot is Available Now")
		}

	case 7:
		// Call the function to view all reservations
		reservations, err := viewAllReservations("Janna Fattouh", ctx)
		if err != nil {
			fmt.Println("Error viewing reservations:", err)
		} else {
			for _, r := range reservations {
				fmt.Println("Name:", r["doctorName"], "Date:", r["date"], "Time:", r["time"])
			}
		}
	}
	//////////////////////////////////////////////////////////////////////////////////////
}
