use presage::{Confirmation, Manager};
use presage_store_sled::SledStore;

pub enum State {
    Started,
    Confirming(Manager<SledStore, Confirmation>),
    Registered,
}

#[derive(Debug, Clone, Copy)]
pub enum Type {
    Primary,
    Secondary,
}

pub enum Registration {
    Unregistered,
    Chosen(Type, State),
}

impl Registration {
    pub fn explain_for_log(&self) -> String {
        match self {
            Self::Unregistered => "No registration started yet.".to_string(),
            Self::Chosen(device, State::Started) => {
                format!("Registration as {device:?} device started.")
            }
            Self::Chosen(device, State::Confirming(_)) => {
                format!("{device:?} device registration is waiting for confirmation.")
            }
            Self::Chosen(device, State::Registered) => {
                format!("Registered as {device:?}.")
            }
        }
    }
}
