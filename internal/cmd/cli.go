package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"PasswordManager/domain"
	"PasswordManager/internal/vault"
)

type CLI struct {
	repo domain.Repository
}

func NewCLI(repo domain.Repository) *CLI {
    return &CLI{repo: repo}
}

func promptSecret(prompt string) (string, error) {
	// Prints initial prompt: "Enter Master Password, or w/e is desired"
	fmt.Print(prompt)

	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}

	fmt.Print()
	return string(bytePassword), nil
}

func (cli *CLI) Execute() error {
	rootCMD := &cobra.Command{
		Use:	"passmgr",
		Short:	"A secure, local password manager",
	}

	addCMD := &cobra.Command{
		Use:	"add [website url] [username]",
		Short:	"Add a new password to the vault",
		Args:	cobra.ExactArgs(2), // fails automatically if there aren't exactly 2 args
		RunE:	func(cmd *cobra.Command, args []string) error {
			website := args[0]
			username := args[1]

			// prompt for master vault password
			masterPw, err := promptSecret("Enter Master Password: ")
			if err != nil {
				return err
			}

			// enter password for target website
			targetPW, err := promptSecret(fmt.Sprintf("Enter new password for %s: ", website))
			if err != nil {
				return err
			}

			//encrypt and save
			encryptedBlob, err := vault.Encrypt(masterPw, targetPW)
			if err != nil {
				return err
			}

			_, err = cli.repo.CreateEntry(cmd.Context(), website, username, encryptedBlob)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully added password for %s!\n", website)
			return nil
		},
	}
	rootCMD.AddCommand(addCMD)

	return rootCMD.Execute()
}