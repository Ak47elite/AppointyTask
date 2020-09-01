package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)





var client *mongo.Client

type Participant struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name string             `json:"name,omitempty" bson:"name,omitempty"`
	Email  string             `json:"email,omitempty" bson:"email,omitempty"`
	RSVP  string             `json:"rsvp,omitempty" bson:"rsvp,omitempty"`
	Meeting  []Meeting             `json:"meeting,omitempty" bson:"meeting,omitempty"`

}

type Meeting struct {
	ID  primitive.ObjectID `json:"_id" bson:"_id"`
	Title string             `json:"title,omitempty" bson:"title"`
	Participants [] Participant  `json:"participant,omitempty" bson:"participant"`
	Start  time.Time  `json:"starttime,omitempty" bson:"starttime"`
	End time.Time      `json:"endtime,omitempty" bson:"endtime"`
	Created time.Time     `bson:",omitempty" json:"created"`

}


func SheduleMeeting(response http.ResponseWriter, request *http.Request)  {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Request-Method","POST")

	var meeting Meeting
	_ = json.NewDecoder(request.Body).Decode(&meeting)
	//participants is not coming
	log.Print(meeting.Participants)
	if meeting.Title=="" || len(meeting.Participants)==0 {
		json.NewEncoder(response).Encode("Please Fill All the details")
		return
	}

//log.Print(meeting.Participants[0])
	for i := 0; i <len(meeting.Participants) ; i++{

		err,_ :=CreateParticipant(meeting.Participants[i],meeting)
		if err!=nil{
			log.Print(err)
			_ = json.NewEncoder(response).Encode("meeting can not be made")
            return
		}
		}
	meet := &Meeting{
		ID: primitive.NewObjectID(),
		Title: meeting.Title,
		Participants: meeting.Participants,
		Start: meeting.Start,
		End: meeting.End,
		Created: time.Now(),
	}
	//if err==nil {
	//	json.NewEncoder(response).Encode("You are already in meeting")
	//	return
	//}
	ctx,_ := context.WithTimeout(context.Background(), 5*time.Second)
	collection := client.Database("mydb").Collection("meeting")

	result,_ := collection.InsertOne(ctx,meet)
	log.Print(result,"logged")
	//var neresult Participant
	var neresult Meeting
	err := collection.FindOne(ctx, bson.M{"_id": result.InsertedID}).Decode(&neresult)
	if err != nil {
		log.Print(err)
		json.NewEncoder(response).Encode("no result found"+err.Error())
		return
	}
	json.NewEncoder(response).Encode(neresult)

}

func CheckRsvp(participant Participant) bool{
	if participant.RSVP == "yes"{
		return false
	}
	return true
}


func CreateParticipant(participant Participant,meeting Meeting) (error,string) {

	if participant.Name=="" || participant.Email=="" || participant.RSVP == ""{
		return errors.New("please fill all the details"), string(0)
	}
	var error Participant
	collection := client.Database("mydb").Collection("participant")

	ctx,_ := context.WithTimeout(context.Background(), 5*time.Second)

	err := collection.FindOne(ctx, bson.M{"email": participant.Email}).Decode(&error)

	var neresult Participant
	err = collection.FindOne(ctx, bson.M{"email": participant.Email}).Decode(&neresult)
	if !CheckRsvp(neresult){
		log.Print("error")
		return errors.New("Meeting can not be made due to participant"), string(0)

	}
	if err != nil {

		participant := &Participant{
			ID: primitive.NewObjectID(),
			Name: participant.Name,
Email: participant.Email,
RSVP: "yes",
Meeting:participant.Meeting,
		}
		result,_ := collection.InsertOne(ctx, participant)
		log.Print(result,"made")
return nil,"success"
	}


	resultUpdate, err := collection.UpdateOne(
		ctx,
		bson.M{"email": participant.Email},
		bson.M{
			"$set": bson.M{
				"rsvp":"yes",

			},
			//"$push":bson.M{
			//
			//},
		},
	)
	log.Print(resultUpdate,"updated")
return nil, "success"
}

func GetMeeting(response http.ResponseWriter, request *http.Request)  {
	response.Header().Set("content-type", "application/json")
	response.Header().Set("Access-Control-Request-Method","GET")
	var meeting Meeting

	log.Print(meeting.Title)
	log.Print(request.URL.Query().Get("id"))
	idUser :=request.URL.Query().Get("id")
	collection := client.Database("mydb").Collection("meeting")

	ctx,_ := context.WithTimeout(context.Background(), 5*time.Second)

	err := collection.FindOne(ctx, bson.M{"_id": idUser}).Decode(&meeting)
	//log.Print(err)
	if err!=nil {
		log.Print(err)
		json.NewEncoder(response).Encode("no result found")
		return
	}

	json.NewEncoder(response).Encode(meeting)

}



func connectMongo(){
	clientOptions := options.Client().ApplyURI("your_mongo_url")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client,_ = mongo.Connect(ctx, clientOptions)
}
func meetings_of_participants(email string)bool{
	if email=="nil" {
		return false
	}
	var meets[] Meeting
	var meeting[] Meeting
	for i:=0;i<len(meeting);i++{
		for j:=0;j<len(meeting[i].Participants);j++{
			if meeting[i].Participants[j].Email==email{
				meets=append(meets,meeting[i])
				break
			}
		}
	}
	fmt.Println(meets)
	return true
}

func MeetingOfParticipant(response http.ResponseWriter, request *http.Request){
	email :=request.URL.Query().Get("email")
	if meetings_of_participants(email){
		json.NewEncoder(response).Encode("meeting Found but not iterated")
	return
	}
	json.NewEncoder(response).Encode("not found")
}
func main() {

	connectMongo()
	fmt.Println("Starting the application...")


	http.HandleFunc("/shedulemeeting",SheduleMeeting)
	http.HandleFunc("/getmeeting", GetMeeting)
	http.HandleFunc("/getparticipantmeeting", MeetingOfParticipant)

	http.ListenAndServe(":3000",nil)
}
