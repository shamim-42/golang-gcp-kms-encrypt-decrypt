# Phone Number Encryption Service

A microservice demonstrating phone number encryption using Google Cloud KMS with auto-rotating keys. This service provides APIs to store encrypted phone numbers in PostgreSQL and retrieve them in decrypted form.

## Prerequisites

- Go 1.21 or later
- Docker
- PostgreSQL client (optional, for direct database access)
- Google Cloud Platform account with KMS enabled
- `make` utility

## Setup

1. **Clone the repository**
```bash
git clone <repository-url>
cd phone-encryption-service
```

2. **Set up environment variables**
```bash
cp .env.example .env
```
Update the `.env` file with your configuration:
```env
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=phone_encryption_db
DB_PORT=54321
PORT=8080
KMS_KEY_PATH=projects/your-project/locations/region/keyRings/your-keyring/cryptoKeys/your-key
GOOGLE_APPLICATION_CREDENTIALS=./your-service-account-key.json
```

3. **Set up GCP Service Account**
- Place your GCP service account key JSON file in the project root
- Update `GOOGLE_APPLICATION_CREDENTIALS` in `.env` to point to your key file
- Ensure the service account has the following roles:
  - `roles/cloudkms.cryptoKeyEncrypterDecrypter`
  - `roles/cloudkms.cryptoKeyViewer`

4. **Install dependencies**
```bash
go mod tidy
```

5. **Install Air (optional, for hot reloading)**
```bash
make -f Makefile.dev install_air
```

## Database Setup

1. **Start PostgreSQL container**
```bash
make dev_postgres
```

2. **Create database**
```bash
make dev_createdb
```

Or, to do everything in one command:
```bash
make dev_reset_db
```

## Running the Service

### Standard Run
```bash
go run main.go
```

### Development Mode (with hot reloading)
```bash
make dev_air
```

## API Endpoints

### Store Phone Number
```bash
curl -X POST http://localhost:8080/phone \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890"}'
```

### Retrieve Phone Number
```bash
curl http://localhost:8080/phone/{id}
```

## Database Access

Connect to the database using any PostgreSQL client:
- Host: localhost
- Port: 54321
- Database: phone_encryption_db
- Username: postgres
- Password: postgres

Example using psql:
```bash
psql -h localhost -p 54321 -U postgres -d phone_encryption_db
```

## Available Make Commands

### Main Commands
- `make dev_postgres` - Start PostgreSQL container
- `make dev_createdb` - Create database
- `make dev_dropdb` - Drop database
- `make dev_reset_db` - Reset entire database environment
- `make dev_air` - Run with hot reloading

### Development Commands
- `make -f Makefile.dev install_air` - Install Air for hot reloading
- `make -f Makefile.dev delete_volume` - Delete Docker volume

## Project Structure
```
.
├── .air.toml           # Air configuration for hot reloading
├── .env                # Environment variables
├── .gitignore         # Git ignore rules
├── Makefile           # Main Makefile
├── Makefile.dev       # Development Makefile
├── README.md          # This file
├── go.mod             # Go modules file
├── go.sum             # Go modules checksum
└── main.go            # Main application code
```

## Security Notes

1. Never commit the service account key to version control
2. Keep your `.env` file secure and never commit it
3. The service uses GCP KMS with auto-rotating keys for enhanced security

## Troubleshooting

1. **Database Connection Issues**
   - Ensure PostgreSQL container is running: `docker ps`
   - Check port availability: `lsof -i :54321`
   - Verify credentials in `.env`

2. **KMS Issues**
   - Verify service account key is present
   - Check KMS_KEY_PATH format
   - Ensure service account has correct permissions

3. **Air Hot Reload Issues**
   - Check tmp directory permissions
   - Verify Air installation: `air -v`
   - Review build errors in `build-errors.log` 