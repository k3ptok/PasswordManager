# passmgr

`passmgr` is a secure, local, and stateless password manager designed specifically for the terminal. It stores all your credentials in an encrypted SQLite vault right on your machine, leveraging robust modern cryptography without relying on external cloud services or background daemons.

⚠️ **CRITICAL DISCLAIMER: THERE IS NO RECOVERY MECHANISM FOR A LOST MASTER PASSWORD.**
Because passmgr is completely stateless and uses your Master Password to derive the encryption keys on the fly, forgetting your password means your vault is permanently inaccessible. You are entirely responsible for remembering your Master Password and backing up your ~/.passmgr/ database folder.

---

## Security Architecture

**Key Derivation**: Argon2id. Protects against GPU brute-forcing and side-channel attacks.

**Encryption**: AES-256-GCM. Provides both data confidentiality and authenticity.

**Storage**: Local SQLite database with CGO dependencies for WINDOWS deployment.

**State:** Stateless execution. The Master Password is never saved to the disk. To prevent vault corruption from typos, the database utilizes an encrypted "canary" string to validate the key before allowing destructive/modifying commands.

**Clipboard**: Passwords copied to your clipboard are automatically wiped after 15 seconds.

## 🛠️ Building from Source (CGO Required)

This project uses the `github.com/mattn/go-sqlite3` driver, which requires CGO and a C compiler to build the database bindings.

### Linux / WSL (Native)
Building the native Linux binary requires the standard GCC compiler.
```bash
# Install GCC if you don't already have it
sudo apt update && sudo apt install gcc

# Build the Linux binary
go build -ldflags="-s -w" -o passmgr main.go
# Clone the repository
git clone https://github.com/k3ptok/passmgr.git
cd passmgr

# Download dependencies
go mod tidy

# Build for your current OS
go build -ldflags="-s -w" -o passmgr main.go

# (Optional) Cross-compile for Windows from Linux/macOS
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o passmgr-win.exe main.go
```
## Usage

### Interactive Shell Mode
```bash
$ passmgr
Interactive passmgr shell started. Type 'help' for commands, or 'exit' to quit.
passmgr> search git
passmgr> get github.com my_user
passmgr> exit
```

**Initialize your vault. Run this first!**
```bash
passmgr> init
```
**Add a new entry with an auto-generated password**
```bash
add github.com my_user --length 24 --no-symbols
```

**Retrieve a password to your clipboard**
```bash
get github.com my_user
```

**Search for a saved service**
```bash
search git
```

**Update an existing entry**
```bash
update-entry github.com my_user
```

**Type `help` to see a full list of commands and flags.**

**IMPORTANT** The passmgr actively refuses duplicates. You will need to curate your .csv files before you upload them. You will lose duplicate records otherwise.

## Data Location

By default, the application will automatically create a hidden directory in your user's home folder to store its configuration, embedded schema migrations, and the database itself.

* **Linux/macOS**: `~/.passmgr/`
* **Windows**: `C:\Users\YourUser\.passmgr\`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. Provided "as is" without warranty of any kind.
