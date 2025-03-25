# SSL Certificate Generation for SecretAI Attest REST Server

This guide provides detailed instructions for generating SSL certificates for the SecretAI Attest REST Server. These certificates enable HTTPS, securing communication between clients and the server.

## Prerequisites

### Install OpenSSL
Ensure that OpenSSL is installed on your system:

```bash
sudo apt update
sudo apt install openssl
```

## Certificate Generation Process

### Step 1: Create Certificate Directory
Create a dedicated directory to store your certificates:

```bash
mkdir -p cert
cd cert
```

### Step 2: Generate a Private Key
Generate a strong 2048-bit RSA private key:

```bash
openssl genrsa -out ssl_key.pem 2048
```

This command creates a private key file named `ssl_key.pem`. The key should be kept secure as it's the foundation of your SSL security.

### Step 3: Generate a Certificate Signing Request (CSR)
Create a Certificate Signing Request using your private key:

```bash
openssl req -new -key ssl_key.pem -out server.csr
```

You'll be prompted to provide information for the certificate:

| Field | Description | Example |
|-------|-------------|---------|
| Country Name (2 letter code) | Your country's ISO code | `US` |
| State or Province | Your state or province | `California` |
| Locality Name | Your city | `San Francisco` |
| Organization Name | Your company name | `SecretAI Inc.` |
| Organizational Unit | Department or team | `Engineering` |
| Common Name | **IMPORTANT**: Server hostname or IP address | `myserver.example.com` or `10.0.0.1` |
| Email Address | Administrative contact | `admin@example.com` |

#### Important Notes About Common Name
- The Common Name **must** match the hostname or IP address used to access the server
- For local testing, use either `localhost` or the server's actual IP address
- If clients will connect using different names, you should consider adding Subject Alternative Names (SAN)

### Step 4: Generate a Self-Signed Certificate
Create a self-signed certificate valid for 365 days:

```bash
openssl x509 -req -days 365 -in server.csr -signkey ssl_key.pem -out ssl_cert.pem
```

This creates the certificate file `ssl_cert.pem`, self-signed with your private key.

### Step 5: Set Proper Permissions
Secure your certificate files with proper permissions:

```bash
chmod 600 ssl_key.pem
chmod 644 ssl_cert.pem
```

This ensures that:
- The private key (`ssl_key.pem`) is readable only by the owner
- The certificate (`ssl_cert.pem`) is readable by everyone but writable only by the owner

### Step 6: Verify the Generated Files
Confirm both files are present and have the correct format:

```bash
ls -l ssl_cert.pem ssl_key.pem
openssl x509 -in ssl_cert.pem -text -noout | head -15
```

The second command displays information about your certificate, including the subject (Common Name) and validity period.

## Integration with SecretAI Attest REST Server

The server automatically looks for certificate files at:
- `cert/ssl_cert.pem`
- `cert/ssl_key.pem`

No additional configuration is needed if you've placed the files in the `cert` directory.

## Security Considerations

### Self-Signed Certificates vs. CA-Signed Certificates
- **Self-signed certificates** are suitable for:
  - Development and testing environments
  - Internal networks where you control all clients
  - Situations where you have a secure method to distribute the certificate to clients

- **CA-signed certificates** are recommended for:
  - Production environments accessible over the internet
  - Systems where clients won't manually trust your certificate
  - Compliance requirements that mandate trusted certificates

### Using Let's Encrypt for Production
For production environments, consider using Let's Encrypt to obtain free, trusted certificates:

1. Install Certbot:
   ```bash
   sudo apt install certbot
   ```

2. Obtain a certificate:
   ```bash
   sudo certbot certonly --standalone -d yourdomain.com
   ```

3. Convert and use the certificates:
   ```bash
   sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem cert/ssl_cert.pem
   sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem cert/ssl_key.pem
   ```

4. Set up automatic renewal:
   ```bash
   sudo systemctl enable certbot.timer
   sudo systemctl start certbot.timer
   ```

## Troubleshooting

### Common Certificate Issues
1. **Certificate not trusted by clients**: Expected with self-signed certificates. Clients will need to explicitly trust the certificate or use `--insecure` flags with tools like curl.

2. **Name mismatch errors**: Occurs when the server hostname doesn't match the certificate's Common Name. Ensure they match exactly.

3. **Permission problems**: If the server can't read the certificate files, check permissions with `ls -l` and adjust if needed.

4. **Certificate expired**: Generate a new certificate with a longer validity period if needed:
   ```bash
   openssl x509 -req -days 3650 -in server.csr -signkey ssl_key.pem -out ssl_cert.pem
   ```
