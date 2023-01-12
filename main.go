package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

var client *iam.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	client = iam.NewFromConfig(cfg)
}

func main() {
	parms := &iam.ListUsersInput{
		MaxItems: aws.Int32(100),
	}

	result, err := client.ListUsers(context.TODO(), parms)

	if err != nil {
		fmt.Println("Error calling iam: ", err)
		return
	}
	total := len(result.Users)
	fmt.Println("Users: ", total)

	people := make([][]string, total)
	count := 0

	for _, user := range result.Users {
		// fmt.Printf("\t%v\n", *user.UserName)

		people[count] = make([]string, 3)

		AccessKeyParams := &iam.ListAccessKeysInput{
			UserName: aws.String(*user.UserName),
		}
		info, err := client.ListAccessKeys(context.TODO(), AccessKeyParams)
		if err != nil {
			fmt.Println("Error calling GetUserName: ", err)
			return

		}
		for _, k := range info.AccessKeyMetadata {
			// fmt.Printf("\t%v\t%v\n", *k.AccessKeyId, *k.CreateDate)
			people[count] = append(people[count], *user.UserName)
			people[count] = append(people[count], *k.AccessKeyId)
			t := *k.CreateDate
			//fmt.Println(t.String())
			people[count] = append(people[count], t.String())
			// fmt.Printf("%v\t%v\t%v\n", *user.UserName, *k.AccessKeyId, *k.CreateDate)
		}
		count++
	}
	for _, user := range people {
		// layout := "2021-10-01 14:32:27 +0000 UTC"
		// t := user
		//k, err := time.Parse(layout, t)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(user)
	}

}
