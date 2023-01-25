package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go/aws/session"
)

var client *iam.Client
var sesClient *ses.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	client = iam.NewFromConfig(cfg)

}

func main() {
	data := getData()
	// fmt.Println(data)
	processData(data)
}

func VerifyEmail(sess *session.Session, email string) error {
	sesClient := ses.New(sess)
	_, err := sesClient.VerifyEmailIdentity(&ses.VerifyEmailIdentityInput{
		EmailAddress: aws.String(email),
	})

	return err
}

func sendEmail() {
	// Define the email message
	email := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String("recipient@example.com"),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Data: aws.String("Hello, this is a test email"),
				},
			},
			Subject: &ses.Content{
				Data: aws.String("Test Email"),
			},
		},
		Source: aws.String("sender@example.com"),
	}

	// Send the email
	req := sesClient.SendEmailRequest(email)
	result, err := req.Send(context.Background())
	if err != nil {
		fmt.Println("Error sending email:", err)
		return
	}

	fmt.Println("Email sent! Message ID:", *result.MessageId)

}

func processData(x string) {
	type Person []struct {
		UserName    string      `json:"UserName"`
		AccessKey   []string    `json:"AccessKey"`
		CreatedDate []time.Time `json:"CreatedDate"`
	}

	CT := time.Now()

	var people Person
	count := 0
	expired := 0
	soon := 0
	notExpired := 0
	total := 0

	json.Unmarshal([]byte(x), &people)

	for _, i := range people {
		total++
		if len(i.AccessKey) >= 2 {
			// fmt.Println(i.UserName, "has more than 1 key")
			for x, y := range i.AccessKey {
				diff := CT.Sub(i.CreatedDate[x])
				if diff > (time.Hour * 2160) {
					fmt.Println("Expired AccessKey", i.UserName, y, i.CreatedDate)
					expired++
				} else if diff < (time.Hour*2160) && diff > (time.Hour*1440) {
					fmt.Println("Almost Expired", i.UserName, y, i.CreatedDate)
					soon++
				} else {
					fmt.Println("All good", i.UserName, y, i.CreatedDate)
					notExpired++
				}
			}
			count++
		} else if len(i.AccessKey) == 1 {
			for x, y := range i.AccessKey {
				diff := CT.Sub(i.CreatedDate[x])
				if diff > (time.Hour * 2160) {
					fmt.Println("Expired AccessKey", i.UserName, y, i.CreatedDate)
					expired++
				} else if diff < (time.Hour*2160) && diff > (time.Hour*1440) {
					fmt.Println("Almost Expired", i.UserName, y, i.CreatedDate)
					soon++
				} else {
					fmt.Println("All good", i.UserName, y, i.CreatedDate)
					notExpired++
				}
			}
		}
	}
	fmt.Println("Users with more than one key: ", count)
	fmt.Println("Users with expired key: ", expired)
	fmt.Println("Users with soon to expire keys: ", soon)
	fmt.Println("Users with no expired keys: ", notExpired)
	fmt.Println("Total User Verification: ", total)

}

func getData() string {

	UserParams := &iam.ListUsersInput{
		MaxItems: aws.Int32(150),
	}
	result, err := client.ListUsers(context.TODO(), UserParams)

	if err != nil {
		fmt.Println("Error calling iam: ", err)
		return ("This failed at calling iam")
	}
	total := len(result.Users)
	fmt.Println("Users: ", total)

	count := 0

	type Person struct {
		UserName    string
		AccessKey   []string
		CreatedDate []time.Time
	}

	var People []Person

	for _, user := range result.Users {
		AccessKeyParams := &iam.ListAccessKeysInput{
			UserName: aws.String(*user.UserName),
		}
		info, err := client.ListAccessKeys(context.TODO(), AccessKeyParams)
		if err != nil {
			fmt.Println("Error calling GetUserName", err)
			return ("This failed at getting the attributes from the username")
		}
		var UN string
		UN = *user.UserName
		// fmt.Printf("%T\n", info.AccessKeyMetadata)

		AK, CD := func(s []types.AccessKeyMetadata) ([]string, []time.Time) {
			var AK []string
			var CD []time.Time
			for _, k := range s {
				// fmt.Printf("%v\t%v\t%v\n", *user.UserName, *k.AccessKeyId, *k.CreateDate)

				//format := "2006-01-02"
				//t := *k.CreateDate
				//t1 := t.String()
				//t2, _ := time.Parse(format, t1)
				AK = append(AK, *k.AccessKeyId)
				CD = append(CD, *k.CreateDate)
			}
			return AK, CD
		}(info.AccessKeyMetadata)

		p1 := Person{
			UserName:    UN,
			AccessKey:   AK,
			CreatedDate: CD,
		}

		People = append(People, p1)

		count++
	}
	data, _ := json.Marshal(People)
	stringData := (string(data))
	return stringData
}
