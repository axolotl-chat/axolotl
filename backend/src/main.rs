mod registering;

#[actix_rt::main]
async fn main() {
    println!("Going to register a user");

    registering::register().await.expect("Failed to register");
}
