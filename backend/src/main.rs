mod registration;
mod requests;

use std::str::FromStr;

use libsignal_service::provisioning::VerificationCodeResponse;
use phonenumber::PhoneNumber;

#[tokio::main]
async fn main() {
    println!("Going to register a user");

    test_registration().await;
}

async fn test_registration() {
    let phonenumber = PhoneNumber::from_str("+4917612345678").expect("Invalid phone number");
    let password = "password";
    let captcha = None;
    let use_voice = false;
    let confirm_code = None;
    let signaling_key = [0u8; 52];

    match confirm_code {
        None => {
            match registration::register_user(&phonenumber, password, captcha, use_voice)
                .await
                .expect("Failed to register user")
            {
                VerificationCodeResponse::Issued => {
                    println!("Registration request was sent");
                }
                VerificationCodeResponse::CaptchaRequired => {
                    println!("Server requires a Captcha. Please generate one and try again!");
                }
            };
        }
        Some(code) => {
            let response = registration::verify_user(&phonenumber, password, code, signaling_key)
                .await
                .expect("Failed to verify user");
            println!("Registered user: {:?}", response);
        }
    }
}
