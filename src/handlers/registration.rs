use presage::{manager::Confirmation, Manager};
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
            Self::Chosen(State::Started) => "Registration started.".to_string(),
            Self::Chosen(State::Confirming(_)) => {
                "Registration is waiting for confirmation.".to_string()
            }
            Self::Chosen(State::Registered) => "Registered.".to_string(),
        }
    }
}
