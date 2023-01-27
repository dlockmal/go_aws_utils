package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamTypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

var client *iam.Client
var sesClient *sesv2.Client

func init() {
	region := os.Getenv("AWS_REGION")
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	client = iam.NewFromConfig(cfg)

	sescfg, createAmazonConfigurationError := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
	)
	if createAmazonConfigurationError != nil {
		panic("configuration error in ses: " + createAmazonConfigurationError.Error())
	}

	sesClient = sesv2.NewFromConfig(sescfg)
}

func sendEmail() {
	recipient := "test"
	sender := "test"
	charset := aws.String("UTF-8")
	contactList := "Idon'tknow"
	subject := "Oh boy... he's testing emails"
	body := "Okay so hear me out, these dividends. They aren't 100% here but this will pay off for a long time"

	// Prepare the email parameters
	emailParams := &sesv2.SendEmailInput{
		Content:                        &types.EmailContent{Simple: &types.Message{Subject: &types.Content{Data: &subject, Charset: charset}, Body: &types.Body{Text: &types.Content{Data: &body}}}},
		ConfigurationSetName:           new(string),
		Destination:                    &types.Destination{ToAddresses: []string{recipient}},
		EmailTags:                      []types.MessageTag{},
		FeedbackForwardingEmailAddress: new(string),
		FeedbackForwardingEmailAddressIdentityArn: new(string),
		FromEmailAddress:            aws.String(sender),
		FromEmailAddressIdentityArn: new(string),
		ListManagementOptions:       &types.ListManagementOptions{ContactListName: &contactList},
		ReplyToAddresses:            []string{},
	}
	_, createMailError := sesClient.SendEmail(context.Background(), emailParams)
	if createMailError != nil {
		panic("Error sending the email: " + createMailError.Error())
	}
}

func main() {
	data := getData()
	// fmt.Println(data)
	processData(data)
	sendEmail()
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

		AK, CD := func(s []iamTypes.AccessKeyMetadata) ([]string, []time.Time) {
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
