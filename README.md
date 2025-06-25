# FormFling - Self-Hosted Contact Form API

A secure, containerized contact form API similar to Formspree that validates origins and sends beautiful emails via SMTP.

## Features

- **Origin validation** - Only accepts requests from specified domains
- **CORS support** - Properly handles cross-origin requests
- **Email validation** - Validates email addresses and input
- **Spam protection** - Basic input sanitization and validation
- **Beautiful email templates** - Professional HTML email formatting similar to Formspree
- **Additional field support** - Automatically includes any extra form fields
- **Docker containerized** - Easy deployment and scaling
- **Health checks** - Built-in health monitoring
- **Configurable** - Environment variable configuration

## Quick Start

1. **Pull the Docker image** from Docker Hub:
   ```bash
   docker pull dungfu/formfling:latest
   ```

2. **Run FormFling** with your configuration:
   ```bash
   docker run -d \
     -p 8080:80 \
     -e TZ="America/New_York" \
     -e SMTP_HOST="smtp.gmail.com" \
     -e SMTP_USERNAME="your-email@gmail.com" \
     -e SMTP_PASSWORD="your-app-password" \
     -e SMTP_FROM_EMAIL="your-email@gmail.com" \
     -e SMTP_TO_EMAIL="your-email@gmail.com" \
     dungfu/formfling:latest
   ```

3. **Test your form endpoint**: `http://localhost:8080/contact.php`

### Alternative: Build from Source

If you prefer to build from source:

1. **Clone or create the project files** in a directory
2. **Configure environment variables** by copying `.env.example` to `.env` and updating values
3. **Build and run** with Docker Compose:

```bash
# Build and start the container
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the container
docker-compose down
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TZ` | Timezone for timestamps | `UTC` |
| `ALLOWED_ORIGINS` | Comma-separated list of allowed origins, or "*" for all | `*` (allows all) |
| `SMTP_HOST` | SMTP server hostname | `smtp.gmail.com` |
| `SMTP_PORT` | SMTP server port | `587` |
| `SMTP_USERNAME` | SMTP username | - |
| `SMTP_PASSWORD` | SMTP password or app password | - |
| `SMTP_FROM_EMAIL` | From email address | - |
| `SMTP_FROM_NAME` | From name | `Contact Form Bot` |
| `SMTP_TO_EMAIL` | Recipient email address | - |
| `SMTP_TO_NAME` | Recipient name | `Website Owner` |

### Gmail Setup (Recommended)

For Gmail SMTP with two-factor authentication:

1. **Enable 2-factor authentication** on your Google Account
2. **Generate an App Password**:
   - Go to [Google Account Settings](https://myaccount.google.com/)
   - Click "Security" → "2-Step Verification" → "App passwords"
   - Select "Mail" → "Other (Custom name)" → Enter "FormFling"
   - Copy the 16-character password (e.g., `abcd efgh ijkl mnop`)
3. **Use the app password** (not your regular password) in `SMTP_PASSWORD`

```yaml
# Gmail configuration example
SMTP_HOST: "smtp.gmail.com"
SMTP_PORT: "587"
SMTP_USERNAME: "your-email@gmail.com"
SMTP_PASSWORD: "your-16-char-app-password"  # Remove spaces when entering
```

## Common Timezone Examples

```bash
# US Timezones
TZ="America/New_York"        # Eastern Time
TZ="America/Chicago"         # Central Time  
TZ="America/Denver"          # Mountain Time
TZ="America/Los_Angeles"     # Pacific Time

# Other Common Timezones
TZ="Europe/London"           # UK
TZ="Europe/Paris"            # Central Europe
TZ="Asia/Tokyo"              # Japan
TZ="Australia/Sydney"        # Australia East Coast
TZ="UTC"                     # Coordinated Universal Time (default)
```

## Usage

### HTML Form Example

```html
<form id="contact-form" action="http://localhost:8080/contact.php" method="POST">
    <input type="text" name="name" placeholder="Your Name" required>
    <input type="email" name="email" placeholder="Your Email" required>
    <input type="text" name="subject" placeholder="Subject" required>
    <textarea name="message" placeholder="Your Message" required></textarea>
    
    <!-- Optional additional fields -->
    <input type="tel" name="phone" placeholder="Phone Number">
    <input type="url" name="website" placeholder="Website">
    <input type="text" name="company" placeholder="Company">
    
    <button type="submit">Send Message</button>
</form>
```

## Email Template

FormFling uses a beautiful HTML email template (`email-template.html`) that creates professional-looking emails similar to Formspree. The template includes:

- **Responsive design** - Works on desktop and mobile email clients
- **Professional styling** - Clean, modern appearance
- **Automatic field inclusion** - Any form fields are automatically formatted and included
- **Technical details section** - IP address, user agent, origin, and referrer information
- **Timestamp** - When the form was submitted

### Customizing the Email Template

You can modify `email-template.html` to match your brand:

- Change colors, fonts, and styling
- Add your logo or branding
- Modify the layout and structure
- Add or remove sections

The template uses these placeholders that are automatically replaced:

- `{{SUBJECT}}` - The form subject
- `{{WEBSITE_NAME}}` - Extracted from the referrer URL
- `{{FORM_FIELDS}}` - Automatically generated form fields
- `{{TIMESTAMP}}` - Submission timestamp
- `{{CLIENT_IP}}` - User's IP address
- `{{USER_AGENT}}` - User's browser information
- `{{ORIGIN}}` - Origin header
- `{{REFERER}}` - Referrer URL
<form id="contact-form" action="http://localhost:8080/contact.php" method="POST">
    <input type="text" name="name" placeholder="Your Name" required>
    <input type="email" name="email" placeholder="Your Email" required>
    <input type="text" name="subject" placeholder="Subject" required>
    <textarea name="message" placeholder="Your Message" required></textarea>
    <button type="submit">Send Message</button>
</form>
```

### JavaScript Example

```javascript
const form = document.getElementById('contact-form');
form.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const formData = new FormData(form);
    
    try {
        const response = await fetch('http://localhost:8080/contact.php', {
            method: 'POST',
            body: formData
        });
        
        const result = await response.json();
        
        if (result.status === 'success') {
            alert('Message sent successfully!');
            form.reset();
        } else {
            alert('Error: ' + result.error);
        }
    } catch (error) {
        alert('Network error: ' + error.message);
    }
});
```

## Security Features

- **Origin Validation**: 
  - By default accepts requests from any domain ("*")
  - Can be restricted to specific domains via `ALLOWED_ORIGINS` environment variable
  - Checks both `Origin` and `Referer` headers when restricted
- **Input Sanitization**: Strips HTML tags and dangerous content
- **Email Validation**: Uses PHP's built-in email validation
- **CORS Control**: Configurable cross-origin request handling
- **Rate Limiting**: Can be added via reverse proxy (nginx, Cloudflare)

## Deployment

### Production Deployment

1. **Use HTTPS**: Ensure your domain uses HTTPS
2. **Environment Variables**: Set production values in `docker-compose.yml`
3. **Reverse Proxy**: Use nginx or Cloudflare for additional security
4. **Monitoring**: Monitor logs and health endpoint

### Docker Hub Deployment

FormFling is automatically built and published to Docker Hub via GitHub Actions when version tags are created.

**Available Tags:**
- `latest` - Latest stable release
- `v1.0.0` - Specific version tags
- `v1.0` - Major.minor version tags
- `v1` - Major version tags

**Note**: Only version tags (e.g., `v1.0.0`) trigger deployments to Docker Hub. Regular commits build but don't deploy.

```bash
# Pull specific version
docker pull dungfu/formfling:v1.0.0

# Pull latest
docker pull dungfu/formfling:latest

# Deploy with Docker Compose using Docker Hub image
version: '3.8'
services:
  formfling:
    image: dungfu/formfling:latest
    ports:
      - "8080:80"
    environment:
      TZ: "America/New_York"
      SMTP_HOST: "smtp.gmail.com"
      SMTP_USERNAME: "your-email@gmail.com"
      SMTP_PASSWORD: "your-app-password"
      SMTP_FROM_EMAIL: "your-email@gmail.com"
      SMTP_TO_EMAIL: "your-email@gmail.com"
```

### Build from Source (Alternative)

```bash
# Build for production
docker build -t your-username/formfling .

# Push to Docker Hub
docker push your-username/formfling

# Deploy on server with origin restrictions
docker run -d \
  -p 8080:80 \
  -e TZ="America/New_York" \
  -e ALLOWED_ORIGINS="https://www.yourdomain1.com,https://yourdomain2.com" \
  -e SMTP_HOST="smtp.gmail.com" \
  -e SMTP_USERNAME="your-email@gmail.com" \
  -e SMTP_PASSWORD="your-app-password" \
  -e SMTP_FROM_EMAIL="your-email@gmail.com" \
  -e SMTP_TO_EMAIL="your-email@gmail.com" \
  your-username/formfling

# Or deploy without origin restrictions (accepts from any domain)
docker run -d \
  -p 8080:80 \
  -e SMTP_HOST="smtp.gmail.com" \
  -e SMTP_USERNAME="your-email@gmail.com" \
  -e SMTP_PASSWORD="your-app-password" \
  -e SMTP_FROM_EMAIL="your-email@gmail.com" \
  -e SMTP_TO_EMAIL="your-email@gmail.com" \
  your-username/formfling
```

## Health Check

The container includes a health check endpoint at `/health.php`:

```bash
curl http://localhost:8080/health.php
```

## Troubleshooting

### Common Issues

1. **Origin errors**: 
   - By default, all origins are allowed ("*")
   - If you set `ALLOWED_ORIGINS`, ensure it matches your website's URL exactly
   - Use comma-separated values for multiple domains
2. **SMTP errors**: Check your email credentials and enable "Less secure app access" or use app passwords
3. **CORS issues**: Verify your domain is in the allowed origins list

### Debugging

```bash
# View container logs
docker-compose logs -f formfling

# Check health status
curl http://localhost:8080/health.php

# Test with curl
curl -X POST http://localhost:8080/contact.php \
  -H "Origin: https://www.yourdomain.com" \
  -d "name=Test&email=test@example.com&subject=Test&message=This is a test message"
```

## License

MIT License - feel free to use and modify FormFling as needed.