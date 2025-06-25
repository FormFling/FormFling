<?php
// Set up headers, timezone, and error reports
header("Content-Type: application/json");
date_default_timezone_set("UTC");
error_reporting(E_ALL ^ E_NOTICE ^ E_DEPRECATED ^ E_STRICT);

// Import PHPMailer classes into the global namespace
use PHPMailer\PHPMailer\PHPMailer;
use PHPMailer\PHPMailer\Exception;

// Load composer's autoloader
require 'vendor/autoload.php';

// Configuration from environment variables
$ALLOWED_ORIGINS = $_ENV['ALLOWED_ORIGINS'] ?? '*';
if ($ALLOWED_ORIGINS !== '*') {
    $ALLOWED_ORIGINS = array_map('trim', explode(',', $ALLOWED_ORIGINS));
}
$SMTP_HOST = $_ENV['SMTP_HOST'] ?? 'smtp.gmail.com';
$SMTP_PORT = $_ENV['SMTP_PORT'] ?? 587;
$SMTP_USERNAME = $_ENV['SMTP_USERNAME'] ?? '';
$SMTP_PASSWORD = $_ENV['SMTP_PASSWORD'] ?? '';
$SMTP_FROM_EMAIL = $_ENV['SMTP_FROM_EMAIL'] ?? '';
$SMTP_FROM_NAME = $_ENV['SMTP_FROM_NAME'] ?? 'FormFling Bot';
$SMTP_TO_EMAIL = $_ENV['SMTP_TO_EMAIL'] ?? '';
$SMTP_TO_NAME = $_ENV['SMTP_TO_NAME'] ?? 'Website Owner';

function died($error) {
    http_response_code(400);
    die(json_encode(["status" => "error", "error" => $error]));
}

function clean_string($string) {
    $bad = array("content-type", "bcc:", "to:", "cc:", "href");
    return str_replace($bad, "", $string);
}

function validateOrigin($allowedOrigins) {
    // If wildcard is set, allow all origins
    if ($allowedOrigins === '*') {
        $origin = $_SERVER['HTTP_ORIGIN'] ?? '*';
        header("Access-Control-Allow-Origin: $origin");
        return true;
    }
    
    // Check Origin header first
    $origin = $_SERVER['HTTP_ORIGIN'] ?? '';
    if ($origin && in_array($origin, $allowedOrigins)) {
        header("Access-Control-Allow-Origin: $origin");
        return true;
    }
    
    // Fallback to Referer header
    $referer = $_SERVER['HTTP_REFERER'] ?? '';
    if ($referer) {
        $refererHost = parse_url($referer, PHP_URL_HOST);
        foreach ($allowedOrigins as $allowedOrigin) {
            $allowedHost = parse_url($allowedOrigin, PHP_URL_HOST);
            if ($refererHost === $allowedHost) {
                header("Access-Control-Allow-Origin: $allowedOrigin");
                return true;
            }
        }
    }
    
    return false;
}

// Handle preflight requests
if ($_SERVER['REQUEST_METHOD'] === 'OPTIONS') {
    if (validateOrigin($ALLOWED_ORIGINS)) {
        header("Access-Control-Allow-Methods: POST, OPTIONS");
        header("Access-Control-Allow-Headers: Content-Type, Origin, Referer");
        header("Access-Control-Max-Age: 86400");
    }
    exit(0);
}

// Validate origin for actual requests
if (!validateOrigin($ALLOWED_ORIGINS)) {
    died("Access denied: Invalid origin");
}

// Only allow POST requests
if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    died("Only POST requests are allowed");
}

// Check if required fields are present
if (!isset($_POST['email']) || !isset($_POST['name']) || !isset($_POST['subject']) || !isset($_POST['message'])) {
    died("Missing required fields");
}

// Sanitize and validate input
$name = clean_string(strip_tags($_POST['name']));
$userEmail = filter_var(clean_string(strip_tags($_POST['email'])), FILTER_SANITIZE_EMAIL);
$userSubject = clean_string(strip_tags($_POST['subject']));
$message = clean_string(strip_tags($_POST['message']));

// Collect additional form fields (optional)
$additionalFields = [];
foreach ($_POST as $key => $value) {
    if (!in_array($key, ['name', 'email', 'subject', 'message']) && !empty($value)) {
        $additionalFields[$key] = clean_string(strip_tags($value));
    }
}

// Validation
$error_message = "";

// Email validation
if (!filter_var($userEmail, FILTER_VALIDATE_EMAIL)) {
    $error_message .= "Invalid email address. ";
}

// Name validation (optional - commented out as in original)
// $string_rgx = "/^[A-Za-z .'-]+$/";
// if (!preg_match($string_rgx, $name)) {
//     $error_message .= "Invalid name format. ";
// }

// Message length validation
if (strlen($message) < 10) {
    $error_message .= "Message too short (minimum 10 characters). ";
}

if (strlen($error_message) > 0) {
    died(trim($error_message));
}

// Get client IP information and create email
$clientIP = $_SERVER['HTTP_X_FORWARDED_FOR'] ?? $_SERVER['HTTP_X_REAL_IP'] ?? $_SERVER['REMOTE_ADDR'] ?? 'unknown';
$userAgent = $_SERVER['HTTP_USER_AGENT'] ?? 'unknown';
$origin = $_SERVER['HTTP_ORIGIN'] ?? 'N/A';
$referer = $_SERVER['HTTP_REFERER'] ?? 'N/A';

// Generate email body using template
$email_body = generateEmailBody([
    'name' => $name,
    'email' => $userEmail,
    'subject' => $userSubject,
    'message' => $message,
    'additional_fields' => $additionalFields,
    'client_ip' => $clientIP,
    'user_agent' => $userAgent,
    'origin' => $origin,
    'referer' => $referer,
    'timestamp' => date("g:i A - j F Y"),
    'website_name' => parse_url($referer, PHP_URL_HOST) ?: 'your website'
]);

function generateEmailBody($data) {
    // Load email template
    $template = file_get_contents(__DIR__ . '/email-template.html');
    
    if (!$template) {
        // Fallback to simple HTML if template not found
        return generateFallbackEmail($data);
    }
    
    // Generate form fields HTML
    $formFields = '';
    
    // Name field
    $formFields .= generateFieldHTML('name', $data['name']);
    
    // Email field  
    $formFields .= generateFieldHTML('email', $data['email']);
    
    // Subject field
    if (!empty($data['subject'])) {
        $formFields .= generateFieldHTML('subject', $data['subject']);
    }
    
    // Message field
    $formFields .= generateFieldHTML('message', $data['message'], true);
    
    // Additional fields
    if (!empty($data['additional_fields'])) {
        foreach ($data['additional_fields'] as $fieldName => $fieldValue) {
            if (!empty($fieldValue)) {
                $formFields .= generateFieldHTML($fieldName, $fieldValue);
            }
        }
    }
    
    // Replace template variables
    $replacements = [
        '{{SUBJECT}}' => htmlspecialchars($data['subject']),
        '{{WEBSITE_NAME}}' => htmlspecialchars($data['website_name']),
        '{{FORM_FIELDS}}' => $formFields,
        '{{TIMESTAMP}}' => htmlspecialchars($data['timestamp']),
        '{{CLIENT_IP}}' => htmlspecialchars($data['client_ip']),
        '{{USER_AGENT}}' => htmlspecialchars($data['user_agent']),
        '{{ORIGIN}}' => htmlspecialchars($data['origin']),
        '{{REFERER}}' => htmlspecialchars($data['referer'])
    ];
    
    return str_replace(array_keys($replacements), array_values($replacements), $template);
}

function generateFieldHTML($label, $value, $isTextarea = false) {
    $displayLabel = ucfirst($label);
    $safeValue = htmlspecialchars($value);
    
    if ($isTextarea) {
        $safeValue = nl2br($safeValue);
    }
    
    return '
    <div style="color:#000000;font-family:\'Open Sans\', \'Helvetica Neue\', Helvetica, Arial, sans-serif;line-height:1.5;padding-top:10px;padding-right:25px;padding-bottom:10px;padding-left:25px;">
        <div class="txtTinyMce-wrapper" style="line-height: 1.5; font-size: 12px; color: #000000; font-family: \'Open Sans\', \'Helvetica Neue\', Helvetica, Arial, sans-serif; mso-line-height-alt: 18px;">
            <p style="margin: 0; font-size: 14px; line-height: 1.5; word-break: break-word; mso-line-height-alt: 21px; margin-top: 0; margin-bottom: 0;">
                <span style="color: #999999;">' . $displayLabel . '</span>
            </p>
            <div style="margin: 8px 0; font-size: 16px; line-height: 1.5; word-break: break-word; mso-line-height-alt: 24px;">
                ' . $safeValue . '
            </div>
        </div>
    </div>';
}

function generateFallbackEmail($data) {
    return '<html><head><meta charset="UTF-8"><style>
    body { margin: 2rem; font-family: Arial, sans-serif; }
    .info { background: #f5f5f5; padding: 1rem; margin: 1rem 0; border-radius: 5px; }
    .message { background: #e8f4f8; padding: 1rem; margin: 1rem 0; border-radius: 5px; }
    .meta { background: #f0f0f0; padding: 0.5rem; margin: 1rem 0; font-size: 0.9em; color: #666; }
    </style></head><body>
    <h2>📧 New Contact Form Submission</h2>
    <div class="info">
        <strong>Name:</strong> ' . htmlspecialchars($data['name']) . '<br>
        <strong>Email:</strong> ' . htmlspecialchars($data['email']) . '<br>
        <strong>Subject:</strong> ' . htmlspecialchars($data['subject']) . '<br>
    </div>
    <div class="message">
        <strong>Message:</strong><br>' . nl2br(htmlspecialchars($data['message'])) . '
    </div>
    <div class="meta">
        <strong>Submitted:</strong> ' . htmlspecialchars($data['timestamp']) . '<br>
        <strong>IP:</strong> ' . htmlspecialchars($data['client_ip']) . '<br>
        <strong>Origin:</strong> ' . htmlspecialchars($data['origin']) . '
    </div>
    </body></html>';
}

// Initialize response array
$page_output = [];

try {
    $mail = new PHPMailer(true);
    
    // Server settings
    $mail->SMTPDebug = 0;
    $mail->CharSet = 'UTF-8';
    $mail->isSMTP();
    $mail->Host = $SMTP_HOST;
    $mail->SMTPAuth = true;
    $mail->Username = $SMTP_USERNAME;
    $mail->Password = $SMTP_PASSWORD;
    $mail->SMTPSecure = PHPMailer::ENCRYPTION_STARTTLS;
    $mail->Port = $SMTP_PORT;
    
    // Recipients
    $mail->setFrom($SMTP_FROM_EMAIL, $SMTP_FROM_NAME);
    $mail->addAddress($SMTP_TO_EMAIL, $SMTP_TO_NAME);
    $mail->addReplyTo($userEmail, $name);
    
    // Content
    $mail->isHTML(true);
    $mail->Subject = '📧 Contact Form: ' . $userSubject;
    $mail->Body = $email_body;
    
    $mail->send();
    $page_output["status"] = "success";
    $page_output["message"] = "Message sent successfully";
    
} catch (Exception $e) {
    error_log("PHPMailer Error: " . $mail->ErrorInfo);
    $page_output["status"] = "error";
    $page_output["error"] = "Failed to send message";
}

header('Content-Type: application/json');
echo json_encode($page_output);
?>