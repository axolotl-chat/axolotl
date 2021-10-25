pub mod registration;
pub mod requests;
pub mod sealedsender;
mod utils;

use self::{
    registration::{register_user, verify_user},
    sealedsender::decrypt_sealed_message,
};
use requests::Request;
use tokio::sync::mpsc;

pub type Queue = mpsc::Sender<Request>;
pub type QueueReceiver = mpsc::Receiver<Request>;

pub struct SignalServiceWrapper {
    queue: QueueReceiver,
    // Put persistent data here
}

impl SignalServiceWrapper {
    pub fn new(queue: QueueReceiver) -> Self {
        // Initialize members here
        Self { queue }
    }

    pub async fn run(mut self) {
        while let Some(req) = self.queue.recv().await {
            self.process(req).await;
        }
    }

    async fn process(&mut self, req: Request) {
        // We ignore the case where the callback receiver has been dropped
        match req {
            Request::Register(data, cb) => {
                let _ = cb.send(register_user(data).await);
            }
            Request::ConfirmRegistration(data, cb) => {
                let _ = cb.send(verify_user(data).await);
            }
            Request::DecryptSealedSender(data, cb) => {
                let _ = cb.send(decrypt_sealed_message(data).await);
            }
        }
    }
}
