package main

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	if newUser.UserType == "patient" {
		patientsCollection := client.Database("Clinic").Collection("Patients")
		_, err = patientsCollection.InsertOne(ctx, bson.M{"name": newUser.Username})
		if err != nil {
			return err
		}
	} else if newUser.UserType == "doctor" {
		doctorsCollection := client.Database("Clinic").Collection("Doctors")
		_, err = doctorsCollection.InsertOne(ctx, bson.M{"name": newUser.Username})
		if err != nil {
			return err
		}
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
		{"isAvailable", "true"},
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
func updateAppointmentDoctor(PatientName string, oldDocName string, newDocName string, selectedSlotdate string, selectedSlotTime string, oldselectedSlotdate string, oldselectedSlotTime string, ctx context.Context) error {

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
	appointmentsCollection := client.Database("Clinic").Collection("Appointments")

	filter := bson.M{
		"name":        newDocName,
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
	// return old slot  to true. make is available to true again

	filterOldApp := bson.M{
		"name":        oldDocName,
		"date":        oldselectedSlotdate,
		"time":        oldselectedSlotTime,
		"isAvailable": "false",
	}
	fmt.Println("Filter:", filterOldApp)
	update = bson.M{
		"$set": bson.M{"isAvailable": "true"},
	}
	updateResult, err = doctorsCollection.UpdateOne(ctx, filterOldApp, update)
	fmt.Println("Update:", update)
	if err != nil {
		fmt.Println("Failed to update slot availability:", err)
		return err
	}
	// update appoitments in appoitments collection
	update = bson.M{
		"$set": bson.M{"doctorName": newDocName, "date": selectedSlotdate, "time": selectedSlotTime},
	}
	updateResult, err = appointmentsCollection.UpdateOne(ctx, bson.M{"patientName": PatientName, "date": oldselectedSlotdate, "time": oldselectedSlotTime}, update)
	if err != nil {
		fmt.Println("Failed to update appointment:", err)
		return err
	}
	if updateResult.ModifiedCount == 0 {
		fmt.Println("No appointment updated.")
		return err
	} else {
		fmt.Println("Appointment updated successfully.")
		return err
	}
	return nil
}

// ///////////////////////////////////////////////////////////////////////////////////////
func updateAppointmentSlot(DocName string, selectedSlotdate string, selectedSlotTime string, oldselectedSlotdate string, oldselectedSlotTime string, ctx context.Context) error {

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

	doctorsCollection := client.Database("Clinic").Collection("Doctors")
	appointmentsCollection := client.Database("Clinic").Collection("Appointments")
	//************************************************************************************************************
	// update new slot in appointment
	update := bson.M{
		"$set": bson.M{"date": selectedSlotdate, "time": selectedSlotTime},
	}
	updateResult, err := appointmentsCollection.UpdateOne(ctx, bson.M{"doctorName": DocName, "date": oldselectedSlotdate, "time": oldselectedSlotTime}, update)
	if err != nil {
		fmt.Println("Failed to update appointment:", err)
		return err
	}
	if updateResult.ModifiedCount == 0 {
		fmt.Println("No appointment updated.")
	} else {
		fmt.Println("Appointment updated successfully.")
	}
	//************************************************************************************************************
	// return old slot of old to true. make is available to true again
	filterNew := bson.M{
		"name":        DocName,
		"date":        oldselectedSlotdate,
		"time":        oldselectedSlotTime,
		"isAvailable": "false",
	}
	fmt.Println("Filter:", filterNew)
	update = bson.M{
		"$set": bson.M{"isAvailable": "true"},
	}
	updateResult, err = doctorsCollection.UpdateOne(ctx, filterNew, update)
	fmt.Println("Update:", update)
	if err != nil {
		fmt.Println("Failed to update slot availability:", err)
		return err
	}
	//*************************************************************************************************************
	// update state of new doc from is available =true to false
	filter := bson.M{
		"name":        DocName,
		"date":        selectedSlotdate,
		"time":        selectedSlotTime,
		"isAvailable": "true",
	}
	update = bson.M{
		"$set": bson.M{"isAvailable": "false"},
	}
	updateResult, err = doctorsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		fmt.Println("Failed to update slot availability:", err)
		return err
	}
	///////////////
	return nil
}

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
			{"$set", bson.M{"isAvailable": "true"}},
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
		var name string
		fmt.Print("Enter the  username : ")
		fmt.Scan(&name)
		var pass string
		fmt.Print("Enter the password : ")
		fmt.Scan(&pass)

		userType, err := signIn(name, pass, ctx)
		if err != nil {
			fmt.Println("Error signing in:", err)
		} else {
			fmt.Println("Signed in as:", userType)
		}

	case 2:
		//sign up
		var name string
		fmt.Print("Enter the  username : ")
		fmt.Scan(&name)
		var pass string
		fmt.Print("Enter the password : ")
		fmt.Scan(&pass)
		var typee string
		fmt.Print("Enter the type : ")
		fmt.Scan(&typee)

		err := signUp(name, pass, typee, ctx)
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
		fmt.Print("enter contact number ")
		var number string
		fmt.Scan(&number)
		fmt.Print("enter date ")
		var datee string
		fmt.Print("enter date ")
		fmt.Scan(&datee)
		var timee string
		fmt.Print("enter time ")
		fmt.Scan(&timee)

		_, err := time.Parse(dateLayout, datee)
		if err != nil {
			fmt.Println("Invalid date format:", err)
			return
		}
		_, err = time.Parse(timeLayout, timee)
		if err != nil {
			fmt.Println("Invalid time format:", err)
			return
		}

		errorr := setSchedule(DocName, datee, timee, number, ctx)
		if errorr != nil {
			fmt.Println("Failed to insert document into Database:", err)
		} else {
			fmt.Println("Data inserted Successfully.")

		}

	case 4:
		//Patients select doctor, view his available slots, then patient chooses a slot.-----------------------------------

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
		var patientName string
		fmt.Print("Enter the patient name : ")
		fmt.Scan(&patientName)
		var dateee string
		fmt.Print("Enter the date : ")
		fmt.Scan(&dateee)

		var timee string
		fmt.Print("Enter the time : ")
		fmt.Scan(&timee)
		// Query the MongoDB collection to find the patient's appointment
		Newfilter := bson.M{
			"patientName": patientName,
			"date":        dateee,
			"time":        timee}

		var Oldappointment struct {
			ID          primitive.ObjectID `bson:"_id,omitempty"`
			Doctor      string             `bson:"doctorName"`
			Date        string             `bson:"date"`
			Time        string             `bson:"time"`
			PatientName string             `bson:"patientName"`
		}
		appointmentsCollection := client.Database("Clinic").Collection("Appointments")

		err = appointmentsCollection.FindOne(ctx, Newfilter).Decode(&Oldappointment)
		if err != nil {
			fmt.Println("Appointment not found:", err)
			return
		}
		fmt.Printf("Your current appointment: Doctor: %s, date: %s, time: %s", Oldappointment.Doctor, Oldappointment.Date, Oldappointment.Time)

		// Ask the patient what they want to update
		fmt.Println("What would you like to update?")
		fmt.Println("1. Doctor")
		fmt.Println("2. Slot")
		fmt.Print("Enter your choice (1 or 2): ")
		var choice int
		fmt.Scan(&choice)
		doctorsCollection := client.Database("Clinic").Collection("Doctors")
		var selectedSlotTime string
		var selectedSlotdate string
		if choice == 1 {
			var newDoctorName string
			fmt.Print("Enter the new doctor's name: ")
			fmt.Scan(&newDoctorName)

			// reserve new slot for new doctor and return all available slot
			// check on available slot of this doctor and return all available slot
			availableSlots, err := getAvailableSlots(ctx, doctorsCollection, newDoctorName)
			if err != nil {
				fmt.Println("Failed to get available slots:", err)
				return
			}

			fmt.Println("Available slots for", newDoctorName, ":")
			for _, slot := range availableSlots {
				fmt.Println(slot)
			}
			//*****************************************************

			fmt.Print("Choose a date&time: ")
			fmt.Scan(&selectedSlotdate, &selectedSlotTime)
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
			errorr := updateAppointmentDoctor(patientName, Oldappointment.Doctor, newDoctorName, selectedSlotdate, selectedSlotTime, Oldappointment.Date, Oldappointment.Time, ctx)
			if errorr != nil {
				fmt.Println("Failed to update the Appointment:", err)
			} else {
				fmt.Println("Appointment updated the doctor Successfully.")
			}
		} else if choice == 2 {
			// find available slot of old doctor then choose it
			availableSlots, err := getAvailableSlots(ctx, doctorsCollection, Oldappointment.Doctor)
			if err != nil {
				fmt.Println("Failed to get available slots:", err)
				return
			}

			fmt.Println("Available slots for", Oldappointment.Doctor, ":")
			for _, slot := range availableSlots {
				fmt.Println(slot)
			}
			fmt.Print("Choose a date&time: ")

			fmt.Scan(&selectedSlotdate, &selectedSlotTime)
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
			//*********************************************************
			// Check if the new slot is available
			if !isSlotAvailable(selectedSlotdate, selectedSlotTime, availableSlots) {
				fmt.Println("Slot is not available.")
				return
			}

			errorr := updateAppointmentSlot(Oldappointment.Doctor, selectedSlotdate, selectedSlotTime, Oldappointment.Date, Oldappointment.Time, ctx)
			if errorr != nil {
				fmt.Println("Failed to update the Appointment:", err)
			} else {
				fmt.Println("Appointment updated the doctor Successfully.")
			}

		} else {
			fmt.Println("Invalid choice. Please enter 1 or 2.")
		}

	case 6:
		// Call the function to cancel appointment
		var patientName string
		fmt.Print("Enter the patient name : ")
		fmt.Scan(&patientName)
		var docName string
		fmt.Print("Enter the doctor name : ")
		fmt.Scan(&docName)
		var datee string
		fmt.Print("Enter the date : ")
		fmt.Scan(&datee)
		var timee string
		fmt.Print("Enter the date : ")
		fmt.Scan(&timee)

		err := cancelAppointment(patientName, docName, datee, timee, ctx)
		if err != nil {
			fmt.Println("Error cancelling appointment:", err)
		} else {
			fmt.Println("This Slot is Available Now")
		}

	case 7:
		// Call the function to view all reservations
		var patientName string
		fmt.Print("Enter the patient Name : ")
		fmt.Scan(&patientName)
		reservations, err := viewAllReservations(patientName, ctx)
		if err != nil {
			fmt.Println("Error viewing reservations:", err)
		} else {
			for _, r := range reservations {
				fmt.Println("Doctor Name:", r["doctorName"], "Date:", r["date"], "Time:", r["time"])
			}
		}
	}
	//////////////////////////////////////////////////////////////////////////////////////
}
