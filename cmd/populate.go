package cmd

import (
	"fmt"

	"errors"
	"log"
	"os"

	"github.com/gobuffalo/pop"
	"github.com/solarwinds/gitlic-check/augit/models"
	swio "github.com/solarwinds/swio-users"
	"github.com/spf13/cobra"
)

// populateCmd represents the populate command
var populateCmd = &cobra.Command{
	Use:   "populate",
	Short: "populate is the command used to populate a local Augit database with users from Azure AD",
	Run: func(cmd *cobra.Command, args []string) {
		err := PopulateDomainUsers()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

type AugitDB struct {
	db *pop.Connection
}

func (adb *AugitDB) Create(inUser *swio.User) error {
	if !inUser.Enabled {
		fmt.Printf("skipping or deleting %s for being disabled\n", inUser.Username)
		// Delete disabled users, so that the offboarding command can check for users existing within the
		// Augit DB associated w/ every GH user
		err := adb.checkForDeletion(inUser)
		if err != nil {
			return err
		}
		return nil
	}
	return adb.upsert(inUser)
}

func (adb *AugitDB) checkForDeletion(inUser *swio.User) error {
	queryUser := &models.GithubUser{}
	err := adb.db.Where("LOWER(email) = LOWER(?)", inUser.Email).First(queryUser)
	if err != nil {
		if models.IsErrRecordNotFound(err) {
			return nil
		}
		return err
	}
	fmt.Printf("deleting %s for disabled or bad email\n", inUser.Email)
	return adb.db.Destroy(queryUser)
}

func (adb *AugitDB) upsert(inUser *swio.User) error {
	return adb.db.Transaction(func(tx *pop.Connection) error {
		foundUser := &models.GithubUser{}
		err := tx.Where("LOWER(email) = LOWER(?)", inUser.Email).First(foundUser)
		if err != nil && !models.IsErrRecordNotFound(err) {
			return err
		} else if err != nil && models.IsErrRecordNotFound(err) {
			ghUser := &models.GithubUser{
				Name:     fmt.Sprintf("%s %s", inUser.FirstName, inUser.LastName),
				Email:    inUser.Email,
				Username: inUser.Username,
			}
			vErrs, err_ := tx.ValidateAndCreate(ghUser)
			if vErrs.HasAny() {
				return vErrs
			}
			if err_ != nil {
				return err_
			}
		} else if err == nil {
			foundUser.Username = inUser.Username
			foundUser.Name = fmt.Sprintf("%s %s", inUser.FirstName, inUser.LastName)
			vErrs, err_ := tx.ValidateAndUpdate(foundUser)
			if vErrs.HasAny() {
				return vErrs
			}
			if err_ != nil {
				return err_
			}
		}
		return nil
	})
}

func PopulateDomainUsers() error {
	cxn, err := pop.Connect(os.Getenv("ENVIRONMENT"))
	if err != nil {
		return err
	}
	id := os.Getenv("AD_CLIENT_ID")
	secret := os.Getenv("AD_SECRET")
	if id == "" || secret == "" {
		return errors.New("must provide id and secret")
	}
	augitDb := &AugitDB{cxn}
	populator := swio.NewPopulator(id, secret, []string{})
	for populator.MoreUsers() {
		users, err := populator.GetUsers()
		if err != nil {
			return err
		}
		for _, user := range users {
			err := augitDb.Create(user)
			if err != nil {
				// TODO: Create error type for array of errors to keep track of failures
				fmt.Printf("[ERROR] skipping user: %+v\n", user)
			}
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(populateCmd)
}
