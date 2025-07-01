# FormFling

![FormFling](https://raw.githubusercontent.com/fireph/FormFling/main/web/static/images/banner-2k-transparent.png)

Haven't you always wanted a contact form on your website, but to keep your email private and not spend a ton of money on a paid service? FormFling is a self-hosted contact form service that forwards submissions to your email via SMTP. Think Formspree, but running on your own server.

[![Tests](https://github.com/fireph/formfling/workflows/Run%20Tests/badge.svg)](https://github.com/fireph/formfling/actions/workflows/test.yml)
[![Coverage](https://github.com/fireph/FormFling/wiki/coverage.svg)](https://github.com/fireph/FormFling/wiki/Test-coverage-report)

## Features

- Single binary deployment
- Gmail SMTP support
- Beautiful HTML email templates
- Origin-based security
- Custom redirect URLs
- Health check endpoint
- reCAPTCHA v3 bot protection
- Docker ready

## Quick Start

### Docker Compose (Recommended)

```bash
git clone https://github.com/FormFling/FormFling.git
cd FormFling
cp docker-compose.yml docker-compose.local.yml
# Edit docker-compose.local.yml with your settings
docker-compose -f docker-compose.local.yml up -d
```

### Docker

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

Configure via environment variables:

**Required:**
- `SMTP_USERNAME` - SMTP Username (Sender email for gmail)
- `SMTP_PASSWORD` - SMTP Password ([setup guide](#gmail-setup))
- `FROM_EMAIL` - Sender email
- `TO_EMAIL` - Where to send form submissions

**Optional:**
- `PORT` - Server port (default: 8080)
- `SMTP_HOST` - SMTP server (default: smtp.gmail.com)
- `SMTP_PORT` - SMTP port (default: 587)
- `ALLOWED_ORIGINS` - Comma-separated allowed domains (default: all)
- `FORM_TITLE` - Form name in emails (default: Contact Me)
- `RECAPTCHA_SITE_KEY` - reCAPTCHA v3 site key
- `RECAPTCHA_SECRET_KEY` - reCAPTCHA v3 secret key
- `RECAPTCHA_MIN_SCORE` - Minimum score threshold (default: 0.5)
- `RECAPTCHA_ACTION` - Expected action name (default: submit)
- `ENABLE_TEST_FORM` - Enable `/test_form` endpoint (default: false)

See [.env.example](.env.example) for all options.

### Gmail Setup

1. Enable 2-Factor Authentication
2. Generate App Password: Google Account → Security → 2-Step Verification → App passwords
3. Use your email as SMTP username and the generated password as SMTP password

## Usage

Add this form to your website:

```html
<form action="https://your-formfling-domain.com/submit" method="POST">
  <input type="text" name="name" required>
  <input type="email" name="email" required>
  <textarea name="message" required></textarea>
  <button type="submit">Send</button>
</form>
```

### With reCAPTCHA v3

```html
<script src="https://www.google.com/recaptcha/api.js"></script>
<form id="contact-form">
  <input type="text" name="name" required>
  <input type="email" name="email" required>
  <textarea name="message" required></textarea>
  <button type="submit"
          class="g-recaptcha" 
          data-sitekey="reCAPTCHA_site_key" 
          data-callback='onSubmit' 
          data-action='submit'>
    Send
  </button>
</form>

 <script>
   function onSubmit(token) {
     document.getElementById("contact-form").submit();
   }
 </script>
```

### JavaScript/AJAX

```javascript
fetch('https://your-formfling-domain.com/submit', {
  method: 'POST',
  body: new FormData(form)
})
.then(response => response.json())
.then(data => {
  if (data.status === 'message sent') {
    // Success
  }
});
```

### AJAX with reCAPTCHA v3

```javascript
document.getElementById('contact-form').addEventListener('submit', function(e) {
  e.preventDefault();
  
  grecaptcha.ready(() => {
    grecaptcha.execute('YOUR_SITE_KEY', {action: 'submit'}).then(token => {
      fetch('https://your-formfling-domain.com/submit', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({
          name: form.name.value,
          email: form.email.value,
          message: form.message.value,
          'g-recaptcha-response': token
        })
      })
      .then(response => response.json())
      .then(data => {
        if (data.status === 'message sent') {
          // Success
        } else {
          // Handle error: data.error
        }
      });
    });
  });
});
```

## API

- `POST /submit` - Submit form
- `GET /health` - Health check
- `GET /status` - Status page
- `GET /test_form` - reCAPTCHA token generator (when `ENABLE_TEST_FORM=true`)

**Response format:**
```json
{"status": "message sent"}
{"status": "error", "error": "description"}
```

## Testing reCAPTCHA

Set `ENABLE_TEST_FORM=true` and visit `/test_form` to generate reCAPTCHA tokens for testing:

```bash
# Set environment variables
export RECAPTCHA_SITE_KEY=your-site-key
export RECAPTCHA_SECRET_KEY=your-secret-key  
export ENABLE_TEST_FORM=true

# Visit http://localhost:8080/test_form in browser
# Copy the generated curl command and run it:
curl -X POST http://localhost:8080/submit \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "message": "Test message",
    "g-recaptcha-response": "generated-token-here"
  }'
```

## Custom Templates

### Custom email template

```bash
docker run -d \
  -v /path/to/template.html:/templates/custom.html \
  -e EMAIL_TEMPLATE=/templates/custom.html \
  dungfu/form-fling:latest
```

Template variables:
- `{{.FormData.Name}}` - Sender's name
- `{{.FormData.Email}}` - Sender's email
- `{{.FormData.Subject}}` - Email subject
- `{{.FormData.Message}}` - Email message
- `{{.FormData.Phone}}` - Sender's phone number
- `{{.FormData.Website}}` - Sender's website
- `{{.SubmittedTime}}` - Time the form was submitted (e.g., "03:04 PM")
- `{{.SubmittedDate}}` - Date the form was submitted (e.g., "02 January 2006")
- `{{.Origin}}` - Origin URL where the form was submitted from

### Custom status template

```bash
docker run -d \
  -v /path/to/template.html:/templates/custom.html \
  -e STATUS_TEMPLATE=/templates/custom.html \
  dungfu/form-fling:latest
```

Template variables:
- `{{.Status}}` - Status type (`success` or `error`)
- `{{.FormTitle}}` - The form title
- `{{.Message}}` - Status message to display
- `{{.RedirectURL}}` - URL to redirect the user back to (if any)

## Development

```bash
git clone https://github.com/fireph/FormFling.git
cd FormFling
go mod download
go run .
```

## Security

- Use HTTPS in production
- Set `ALLOWED_ORIGINS` to restrict access
- Use Gmail App Passwords
- Consider rate limiting at proxy level

## License

MIT License - see [LICENSE](LICENSE) file.
