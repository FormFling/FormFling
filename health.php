<?php
// Simple health check endpoint
header('Content-Type: application/json');

$health = [
    'status' => 'healthy',
    'timestamp' => date('c'),
    'service' => 'formfling-api',
    'version' => '1.0.0'
];

// Check if PHPMailer is available
try {
    require 'vendor/autoload.php';
    use PHPMailer\PHPMailer\PHPMailer;
    $mail = new PHPMailer();
    $health['phpmailer'] = 'available';
} catch (Exception $e) {
    $health['phpmailer'] = 'error';
    $health['status'] = 'unhealthy';
}

echo json_encode($health);
?>