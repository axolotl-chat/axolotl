use super::Result;
use crate::config::SignalConfig;
use crate::store::Storage;
use crate::store::StorageLocation;
use libsignal_service::cipher::ServiceCipher;
use libsignal_service::configuration::ServiceConfiguration;
use libsignal_service::configuration::SignalServers;
use libsignal_service::prelude::Envelope;
// use protocol::envelope::*;
use serde::{Deserialize, Serialize};

#[derive(Deserialize, Debug)]
pub struct SealedSenderMessage {
    #[serde(with = "serde_str")]
    pub uuid: String,
    pub message:[u8; 32],
}

#[derive(Serialize)]
pub struct DecryptSealedMessageResponse {
    pub message: [u8; 32],
}
fn service_cfg() -> ServiceConfiguration {
    // XXX: read the configuration files!
    SignalServers::Production.into()
}
pub async fn decryptSealedMessage(
    data: SealedSenderMessage,
) -> Result<DecryptSealedMessageResponse> {
    let service_cfg = service_cfg();
    let config = SignalConfig::default();
    let storage = open_storage(&config).await?;

    let mut cipher = ServiceCipher::new(
        storage.clone(),
        storage.clone(),
        storage.clone(),
        storage.clone(),
        rand::thread_rng(),
        service_cfg.credentials_validator().expect("trust root"),
    );
    let signaling_key = storage.signaling_key().await?;
    let ret = Envelope::decrypt(&data.message, &signaling_key, true).unwrap();
    // let ret = Envelope::try_from(data.message).unwrap();
    // let ret : Envelope =  protobuf::Message::parse_from_bytes(data.message).unwrap();
    cipher.open_envelope(ret);

    Ok(DecryptSealedMessageResponse { message: [0; 32] })
}

async fn open_storage(config: &crate::config::SignalConfig) -> anyhow::Result<Storage> {
    let home = dirs::home_dir().unwrap()
    .join(".local")
    .join("share")
    .join("textsecure.nanuc");
    let location = StorageLocation::Path(home);
    let storage = Storage::open(&location).await?;
    Ok(storage)
}
