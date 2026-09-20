package cmd

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"github.com/atotto/clipboard"

	"PasswordManager/domain"
	"PasswordManager/internal/vault"
	"PasswordManager/config"
)

type CLI struct {
	repo domain.Repository
	cfg *config.AppConfig
}

func NewCLI(repo domain.Repository, cfg *config.AppConfig) *CLI {
    return &CLI{
		repo: 	repo,
		cfg:	cfg,
	}
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

func verifyMasterPassword(ctx context.Context, repo domain.Repository, masterPW string) error {
	encryptedCanary, err := repo.GetConfig(ctx, "canaery")
	if err != nil {
		
		return fmt.Errorf("Vault not initialized, or is corrupted. Run 'pssmgr init' first.")
	}

	decrypted, err := vault.Decrypt(masterPW, encryptedCanary)
	if err != nil || decrypted != "AUTH_OK" {
		return fmt.Errorf("Incorrect master password.")
	}
	return nil
}

func (cli *CLI) Execute() error {
	if len(os.Args) > 1 {
		cli.buildCommands().Execute()
	}

	fmt.Println("passmgr started. Type 'help' for commands, or 'exit' to quit")
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("passmgr > ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			break
		}

		args := strings.Fields(input)
		cmd := cli.buildCommands()
		cmd.SetArgs(args)

		if err := cmd.Execute(); err != nil {
			//prevent crashes with bad inputs
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		slog.Error("REPL error: Scanner broke", "error", err)
	}

	return nil
}

func (cli *CLI) buildCommands() *cobra.Command {
	rootCMD := &cobra.Command{
	Use:	"passmgr",
	Short:	"A secure, local password manager",
	Long:	"passmgr is a stateless, local password manager designed for the terminal.\n" +
	"It uses Argon2id for key derivation and AES-GCM for authenticated encryption.\n" +
	"All data is stored locally in an encrypted SQLite vault.\n",
	Example: "init\n" +
  		"add github.com my_user -l 24 -s\n" +
  		"get github.com my_user\n" +
  		"search git\n",
	SilenceUsage: true,
	}
	
	var addLength int
	var addNoSymbols bool
	addCMD := &cobra.Command{
		Use:	"add [website url] [username]",
		Short:	"Add a new password to the vault",
		Args:	cobra.ExactArgs(2), // fails automatically if there aren't exactly 2 args
		RunE:	func(cmd *cobra.Command, args []string) error {
			website, username := args[0], args[1]

			// prompt for master vault password
			masterPw, err := promptSecret("Enter Master Password: ")
			if err != nil {
				return err
			}

			if err := verifyMasterPassword(cmd.Context(), cli.repo, masterPw); err != nil {
				return err
			}

			charType := "with symbols"
			if addNoSymbols {
				charType = "NO symbols"
			}

			// enter password for target website
			targetPW, err := promptSecret(fmt.Sprintf("Enter new password for %s (leave blank to auto-generate %d chars, %s): ", website, addLength, charType))
			if err != nil {
				return err
			}

			var wasGenerated bool
			if targetPW == "" {
				targetPW, err = vault.GeneratePassword(addLength, !addNoSymbols)
				if err != nil {
					slog.Error("Password generation failed", "error", err)
					return fmt.Errorf("Failed to generate password: %w", err)
				}
				wasGenerated = true

				if err := clipboard.WriteAll(targetPW); err != nil {
					return fmt.Errorf("Failed to write to clipboard: %w", err)
				}

				fmt.Printf("\n Auto-generated %d-character password copied to clipboard\n", addLength)
				fmt.Printf("\nYou may continue registering your account at %s\n", website)
			}

			fmt.Println("Press [ENTER] to save to vault, or type 'abort' to cancel entry registration: ")

			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				if wasGenerated { clipboard.WriteAll("") }
				return err
			}

			input = strings.TrimSpace(strings.ToLower(input))
			if input == "abort" || input == "abort-entry" {
				if wasGenerated { clipboard.WriteAll("") }
				fmt.Println("\nAborted. Clipboard cleared and entry discarded.")
				return nil
			}
			

			//encrypt and save
			encryptedBlob, err := vault.Encrypt(masterPw, targetPW)
			if err != nil {
				if wasGenerated { clipboard.WriteAll("") }
				return err
			}

			_, err = cli.repo.CreateEntry(cmd.Context(), website, username, encryptedBlob)
			if err != nil {
				if wasGenerated { clipboard.WriteAll("") }
				return err
			}

			fmt.Printf("Successfully added password for %s!\n", website)
			if wasGenerated {
				clipboard.WriteAll("")
				fmt.Println("Clipboard cleared")
			}
			return nil
		},
	}

	getCMD := &cobra.Command{
		Use:	"get [website] [username]",
		Short:	"Retrieve a password",
		Args:	cobra.ExactArgs(2),
		RunE:	func(cmd *cobra.Command, args []string) error {
			website, username := args[0], args[1]
			
			entry, err := cli.repo.GetEntry(cmd.Context(), website, username)
			if err != nil {
				return err
			}

			masterPW, err := promptSecret("Enter Master Password: ")
			if err != nil {
				return err
			}

			if err := verifyMasterPassword(cmd.Context(), cli.repo, masterPW); err != nil {
				return err
			}

			decrypted, err := vault.Decrypt(masterPW, entry.EncryptedPassword)
			if err != nil {
				return fmt.Errorf("decryption failed - incorrect master password or corrupted data")
			}

			if err := clipboard.WriteAll(decrypted); err != nil {
				return fmt.Errorf("Failed to write to clipboard: %w", err)
			}

			fmt.Printf("\n Password for %s (%s) copied to clipboard\n", website, username)
			fmt.Println("Clipboard will clear in 15 seconds (CTRL + C will exit program and clear clipboard)...")

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

				
			select {
			case <-time.After(15 * time.Second):
					
			case <-sigChan:
				fmt.Println("\nInterrupt received. Clearing early...")
			}

				
			clipboard.WriteAll("")
			fmt.Println("Clipboard cleared.")
			return nil
		},
	}

	listCMD := &cobra.Command{
		Use:	"list-all",
		Short:	"List all saved websites and usernames",
		Args:	cobra.NoArgs,
		RunE:	func(cmd *cobra.Command, args []string) error {
			entries, err := cli.repo.ListEntries(cmd.Context())
			if err != nil {
				return err
			}

			fmt.Println("Saved Entries:")
			for _, entry := range entries {
				fmt.Printf("- %s (User: %s)", entry.WebsiteURL, entry.Username)
			}
			return nil

		},

	}

	var updateLength int
	var updateNoSymbols bool
	updateCMD := &cobra.Command{
		Use:	"update-entry [website] [username]",
		Short:	"Update an existing password",
		Args:	cobra.ExactArgs(2),
		RunE:	func(cmd *cobra.Command, args []string) error {
			website, username := args[0], args[1]

			// verify entry exists
			_, err := cli.repo.GetEntry(cmd.Context(), website, username)
			if err != nil {
				return err
			}

			masterPW, err := promptSecret("Enter master password: ")
			if err != nil {
				return err
			}

			if err := verifyMasterPassword(cmd.Context(), cli.repo, masterPW); err != nil {
				return err
			}

			charType := "with symbols"
			if updateNoSymbols {
				charType = "NO symbols"
			}
			newPW, err := promptSecret(fmt.Sprintf("Enter NEW password for %s (leave blank to auto-generate %d chars, %s)", website, updateLength, charType))
			if err != nil {
				return err
			}

			var wasGenerated bool

			if newPW == "" {
				newPW, err := vault.GeneratePassword(updateLength, !updateNoSymbols)
				if err != nil {
					return fmt.Errorf("Failed to generate password")
				}
				wasGenerated = true

				if err := clipboard.WriteAll(newPW); err != nil {
					return fmt.Errorf("failed to copy to clipboard: %w", err)
				}

				fmt.Printf("\n Auto-generated %d-character password copied to clipboard\n", updateLength)
				fmt.Printf("\nYou may continue registering your account at %s\n", website)
			}

			fmt.Printf("\nReady to update entry for %s (%s).\n", website, username)
			fmt.Print("Press [ENTER] to save to vault, or type 'abort' to cancel: ")

			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				if wasGenerated { clipboard.WriteAll("") }
				return err
			}

			input = strings.TrimSpace(strings.ToLower(input))
			if input == "abort" || input == "abort-entry" {
				fmt.Println("\n Aborted. Clipboard cleared and entry unchanged")
				return nil
			}

			encryptedBlob, err := vault.Encrypt(masterPW, newPW)
			if err != nil {
				if wasGenerated { clipboard.WriteAll("") }
				return err
			}

			if err := cli.repo.UpdateEntry(cmd.Context(), website, username, encryptedBlob); err != nil {
				if wasGenerated { clipboard.WriteAll("") }
				return err
			}

			fmt.Printf("Successfully updated password for %s", website)

			if wasGenerated {
				clipboard.WriteAll("")
				fmt.Println("Clipboard cleared.")
				}
			return nil

		},
		
	}
	updateCMD.Flags().IntVarP(&updateLength, "length", "l", 32, "Length of the auto-generated password")
	updateCMD.Flags().BoolVarP(&updateNoSymbols, "no-symbols", "s", false, "Exclude special characters if auto-generating")

	searchCmd := &cobra.Command{
			Use:   "search [keyword]",
			Short: "Search for a saved service",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				keyword := args[0]
				
				entries, err := cli.repo.SearchEntries(cmd.Context(), keyword)
				if err != nil {
					return err
				}
				
				if len(entries) == 0 {
					fmt.Printf("No entries found matching '%s'\n", keyword)
					return nil
				}

				fmt.Printf("Found %d entries matching '%s':\n", len(entries), keyword)
				for _, e := range entries {
					fmt.Printf("- %s (User: %s)\n", e.WebsiteURL, e.Username)
				}
				return nil
			},
		}

	importCmd := &cobra.Command{
		Use:   "import-csv [file_path]",
		Short: "Bulk import passwords from a CSV file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("could not open file: %w", err)
			}
			defer file.Close()

			reader := csv.NewReader(file)
			records, err := reader.ReadAll()
			if err != nil {
				return fmt.Errorf("could not parse CSV: %w", err)
			}

			if len(records) < 2 {
				return fmt.Errorf("CSV appears to be empty or missing data rows")
			}

			masterPw, err := promptSecret("Enter Master Password: ")
			if err != nil {
				return err
			}

			if err := verifyMasterPassword(cmd.Context(), cli.repo, masterPw); err != nil {
				return err
			}

			successCount := 0
			fmt.Println("Importing entries...")

			for i := 1; i < len(records); i++ {
				row := records[i]
				if len(row) < 3 {
					fmt.Printf("Skipping row %d: insufficient columns\n", i+1)
					continue
				}

				website := strings.TrimSpace(row[0])
				username := strings.TrimSpace(row[1])
				plainPassword := strings.TrimSpace(row[2])

				if website == "" || plainPassword == "" {
					fmt.Printf("Skipping row %d: missing service or password\n", i+1)
					continue
				}

				encryptedBlob, err := vault.Encrypt(masterPw, plainPassword)
				if err != nil {
					fmt.Printf("Failed to encrypt %s: %v\n", website, err)
					continue
				}

				if err := cli.repo.ImportEntry(cmd.Context(), website, username, encryptedBlob); err != nil {
					fmt.Printf("Failed to save %s: %v\n", website, err)
					continue
				}
				successCount++
			}

			fmt.Printf("\n Import complete! Successfully processed %d entries.\n", successCount)
			return nil
		},
	}

	deleteCMD := &cobra.Command{
		Use:	"delete-entry [website] [username]",
		Short:	"Delete a saved entry and it's password",
		Args:	cobra.ExactArgs(2),
		RunE:	func(cmd *cobra.Command, args []string) error {
			website, username := args[0], args[1]

			if err := cli.repo.DeleteEntry(cmd.Context(), website, username); err != nil {
				return err
			}
			fmt.Printf("Successfully deleted entry for %s", website)
			return nil
		},
	}

	initCMD := &cobra.Command {
		Use:	"init",
		Short:	"Initialize the password manager and set the Master Password",
		Args:	cobra.NoArgs,
		RunE:	func(cmd *cobra.Command, args []string) error {
			_, err := cli.repo.GetConfig(cmd.Context(), "canary")
			if err == nil {
				return fmt.Errorf("Vault already initialized: %w", err)
			}

			fmt.Println("Warning: If you lose this Master Password, your vault is permanently gone.")
			masterPW, err := promptSecret("\nEnter Master Password: \n")
			if err != nil {
				return err
			}

			confirmPW, err := promptSecret("\nConfirm Master Password: \n")
			if err != nil {
				return err
			}

			if masterPW != confirmPW {
				return fmt.Errorf("Entered passwords do not match, aborting initialization...")
			}

			encryptedCanary, err := vault.Encrypt(masterPW, "AUTH_OK")
			if err != nil {
				return err
			}

			if err := cli.repo.SetConfig(cmd.Context(), "canary", encryptedCanary); err != nil {
				return err
			}

			fmt.Println("Vault initialized successfully.")
			return nil
		},
	}

	backupCmd := &cobra.Command{
			Use:   "backup [destination_path]",
			Short: "Create a secure backup of your encrypted vault",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				destPath := args[0]

				// cfg.DBPath comes from your config.Load() in main.go. 
				// You can pass it into NewCLI if you need access to it here.
				sourceFile, err := os.Open(cli.cfg.DBPath)
				if err != nil {
					return fmt.Errorf("failed to open vault: %w", err)
				}
				defer sourceFile.Close()

				destFile, err := os.Create(destPath)
				if err != nil {
					return fmt.Errorf("failed to create backup file: %w", err)
				}
				defer destFile.Close()

				if _, err := io.Copy(destFile, sourceFile); err != nil {
					return fmt.Errorf("failed to write backup: %w", err)
				}

				// Enforce strict permissions on the backup file
				if err := os.Chmod(destPath, 0600); err != nil {
					fmt.Printf("Warning: Could not set strict permissions on backup: %v\n", err)
				}

				fmt.Printf(" Vault successfully backed up to %s\n", destPath)
				return nil
			},
		}
	rootCMD.CompletionOptions.DisableDefaultCmd = true
	rootCMD.AddCommand(addCMD, getCMD, listCMD, updateCMD, deleteCMD, initCMD, backupCmd,searchCmd, importCmd)
	
	return rootCMD
}