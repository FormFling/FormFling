![FormFling](https://raw.githubusercontent.com/fireph/FormFling/main/images/banner-2k-transparent.png)

A self-hosted, lightweight form submission service written in Go, designed as an alternative to Formspree. FormFling allows you to handle contact form submissions on static websites by forwarding them via email using SMTP.

[![Tests](https://github.com/fireph/formfling/workflows/Run%20Tests/badge.svg)](https://github.com/fireph/formfling/actions/workflows/test.yml)
[![Coverage](https://github.com/fireph/FormFling/wiki/coverage.svg)](https://github.com/fireph/FormFling/wiki/Test-coverage-report)

## Features

- 🚀 **Lightweight**: Single binary with minimal resource usage
- 🔒 **Origin Protection**: Configurable allowed origins to prevent unauthorized usage
- 📧 **Gmail Support**: Built-in support for Gmail SMTP
- 🎨 **Beautiful Emails**: Responsive HTML email templates matching Formspree's design
- 🐳 **Docker Ready**: Easy deployment with Docker and Docker Compose
- 🔍 **Health Checks**: Built-in health endpoint for monitoring
- 🛡️ **Security**: Input validation and sanitization

## Quick Start

### Using Docker Compose

1. Clone this repository:
```bash
git clone https://github.com/fireph/FormFling.git
cd FormFling
```

2. Copy and edit the docker-compose.yml file with your settings:
```bash
cp docker-compose.yml docker-compose.local.yml
# Edit docker-compose.local.yml with your email settings
```

3. Run with Docker Compose:
```bash
docker-compose -f docker-compose.local.yml up -d
```

### Using Docker

```bash
docker run -d \
  --name formfling \
  -p 8080:8080 \
  -e SMTP_USERNAME=your-email@gmail.com \
  -e SMTP_PASSWORD=your-app-password \
  -e FROM_EMAIL=your-email@gmail.com \
  -e TO_EMAIL=recipient@example.com \
  -e ALLOWED_ORIGINS=https://www.yourdomain.com \
  dungfu/form-fling:latest
```

## Configuration

FormFling is configured entirely through environment variables:

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `SMTP_USERNAME` | SMTP username (usually your email) | `your-email@gmail.com` |
| `SMTP_PASSWORD` | SMTP password (Gmail App Password) | `your-app-password` |
| `FROM_EMAIL` | Sender email address | `your-email@gmail.com` |
| `TO_EMAIL` | Recipient email address | `recipient@example.com` |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `SMTP_HOST` | SMTP server hostname | `smtp.gmail.com` |
| `SMTP_PORT` | SMTP server port | `587` |
| `FROM_NAME` | Sender display name | `FormFling` |
| `TO_NAME` | Recipient display name | `` |
| `ALLOWED_ORIGINS` | Comma-separated allowed origins | `*` (allows all) |
| `FORM_TITLE` | Form title in emails | `Contact Me` |
| `EMAIL_TEMPLATE` | Path to custom email template file | `./email_template.html` |
| `STATUS_TEMPLATE` | Path to custom status page template file | `./status_template.html` |

### Gmail Setup

1. Enable 2-Factor Authentication on your Google account
2. Generate an App Password:
   - Go to Google Account settings
   - Security → 2-Step Verification → App passwords
   - Generate a new app password for "Mail"
3. Use your email and the generated app password in the configuration

### Custom Email Templates

FormFling supports custom email templates. You can provide your own HTML template by setting the `EMAIL_TEMPLATE` environment variable to the path of your template file.

#### Using Docker with Custom Template

```bash
docker run -d \
  --name formfling \
  -p 8080:8080 \
  -v /path/to/your/template:/templates/custom_email_template.html \
  -e SMTP_USERNAME=your-email@gmail.com \
  -e SMTP_PASSWORD=your-app-password \
  -e FROM_EMAIL=your-email@gmail.com \
  -e TO_EMAIL=recipient@example.com \
  -e EMAIL_TEMPLATE=/templates/custom_email_template.html \
  dungfu/form-fling:latest
```

#### Template Variables

Your custom template can use the following variables:
- `{{.FormData.Name}}` - Sender's name
- `{{.FormData.Email}}` - Sender's email
- `{{.FormData.Subject}}` - Email subject
- `{{.FormData.Message}}` - Email message
- `{{.FormData.Phone}}` - Sender's phone number
- `{{.FormData.Website}}` - Sender's website
- `{{.SubmittedTime}}` - Time the form was submitted (e.g., "03:04 PM")
- `{{.SubmittedDate}}` - Date the form was submitted (e.g., "02 January 2006")
- `{{.Origin}}` - Origin URL where the form was submitted from

### Custom Status Page Templates

FormFling supports custom status page templates. You can provide your own HTML template by setting the `STATUS_TEMPLATE` environment variable to the path of your template file.

#### Using Docker with Custom Status Template

```bash
docker run -d \
  --name formfling \
  -p 8080:8080 \
  -v /path/to/your/status_template:/templates/custom_status_template.html \
  -e SMTP_USERNAME=your-email@gmail.com \
  -e SMTP_PASSWORD=your-app-password \
  -e FROM_EMAIL=your-email@gmail.com \
  -e TO_EMAIL=recipient@example.com \
  -e STATUS_TEMPLATE=/templates/custom_status_template.html \
  dungfu/form-fling:latest
```

#### Status Template Variables

Your custom status template can use the following variables:
- `{{.Status}}` - Status type (`success` or `error`)
- `{{.FormTitle}}` - The form title
- `{{.Message}}` - Status message to display
- `{{.RedirectURL}}` - URL to redirect the user back to (if any)

## HTML Form Integration

Create a form on your static website that submits to your FormFling instance:

```html
<form action="https://your-formfling-domain.com/submit" method="POST">
  <input type="text" name="name" placeholder="Your Name" required>
  <input type="email" name="email" placeholder="Your Email" required>
  <input type="text" name="subject" placeholder="Subject">
  <input type="tel" name="phone" placeholder="Phone (optional)">
  <input type="url" name="website" placeholder="Website (optional)">
  <textarea name="message" placeholder="Your Message" required></textarea>
  <button type="submit">Send Message</button>
</form>
```

### JavaScript Integration

For better user experience, handle the form submission with JavaScript:

```javascript
document.getElementById('contact-form').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const formData = new FormData(this);
    
    try {
        const response = await fetch('https://your-formfling-domain.com/submit', {
            method: 'POST',
            body: formData
        });
        
        const result = await response.json();
        
        if (result.status === 'message sent') {
            alert('Message sent successfully!');
            this.reset();
        } else {
            alert('Error: ' + result.error);
        }
    } catch (error) {
        alert('Network error. Please try again.');
    }
});
```

## API Endpoints

- `POST /submit` - Submit form data
- `GET /health` - Health check endpoint

### Response Format

Success:
```json
{
  "status": "message sent"
}
```

Error:
```json
{
  "status": "error",
  "error": "error description"
}
```

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/fireph/FormFling.git
cd FormFling

# Install dependencies
go mod download

# Build
go build -o formfling .

# Run
./formfling
```

### Local Development

```bash
# Set environment variables
export SMTP_USERNAME=your-email@gmail.com
export SMTP_PASSWORD=your-app-password
export FROM_EMAIL=your-email@gmail.com
export TO_EMAIL=recipient@example.com
export ALLOWED_ORIGINS=http://localhost:3000

# Run
go run .
```

## Deployment

### GitHub Actions

The repository includes a GitHub Actions workflow that automatically builds and pushes Docker images to Docker Hub when you push to the main branch or create a release.

1. Set up Docker Hub secrets in your GitHub repository:
   - `DOCKERHUB_USERNAME` - Your Docker Hub username
   - `DOCKERHUB_TOKEN` - Your Docker Hub access token

2. Push to main branch or create a release tag

### Manual Deployment

1. Build and push the Docker image:
```bash
docker build -t dungfu/form-fling:latest .
docker push dungfu/form-fling:latest
```

2. Deploy to your server using docker-compose or your preferred orchestration tool

## Security Considerations

- Always use HTTPS in production
- Set `ALLOWED_ORIGINS` to restrict form submissions to your domains
- Use Gmail App Passwords instead of your main password
- Consider rate limiting at the reverse proxy level
- Keep your FormFling instance updated

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by [Formspree](https://formspree.io/)
