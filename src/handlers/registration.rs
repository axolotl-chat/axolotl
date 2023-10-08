use presage::{Confirmation, Manager};
use presage_store_sled::SledStore;

pub enum State {
    Started,
    Confirming(Manager<SledStore, Confirmation>),
    Registered,
}

pub enum Registration {
    Unregistered,
    Chosen(State),
}

impl Registration {
    pub fn explain_for_log(&self) -> String {
        match self {
            Self::Unregistered => "No registration started yet.".to_string(),
            Self::Chosen(State::Started) => {
                format!("Registration started.")
            }
            Self::Chosen(State::Confirming(_)) => {
                format!("Registration is waiting for confirmation.")
            }
            Self::Chosen(State::Registered) => {
                format!("Registered.")
            }
        }
    }
}
