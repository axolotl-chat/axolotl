use presage as p;
use presage_store_sled::SledStoreError;

const FAILED_TO_LOOK_UP_ADDRESS: &str = "failed to lookup address information";
type PresageError = presage::Error<presage_store_sled::SledStoreError>;

#[derive(Debug)]
pub enum ApplicationError {
    ManagerThreadPanic,
    NoInternet,
    Presage(presage::Error<presage_store_sled::SledStoreError>),
    SledStore(presage_store_sled::SledStoreError),
    UnauthorizedSignal,
    SendFailed(presage::libsignal_service::sender::MessageSenderError),
    ReceiveFailed(libsignal_service::receiver::MessageReceiverError),
    WebSocketError,
    WebSocketHandleMessageError(String),
    RegistrationError(String),
    InvalidRequest,
    RegistrationSuccesful,
}

impl std::fmt::Display for ApplicationError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ApplicationError::ManagerThreadPanic => writeln!(
                f,
                "{}",
                "A part of the application crashed."
            ),
            ApplicationError::NoInternet => writeln!(
                f,
                "{}",
                "There does not seem to be a connection to the internet available."
            ),

            ApplicationError::UnauthorizedSignal => writeln!(
                f,
                "{}",
                "You do not seem to be authorized with Signal. Please delete the database and relink the application."
            ),
            ApplicationError::SendFailed(_) => writeln!(
                f,
                "{}",
                "Sending a message failed."
            ),
            ApplicationError::ReceiveFailed(_) => writeln!(
                f,
                "{}",
                "Receiving a message failed."
            ),
            ApplicationError::Presage(e) => writeln!(
                f,
                "{}: {}",
                "presage error: ",
                e.to_string()
            ),
            ApplicationError::WebSocketError => writeln!(
                f,
                "{}",
                "The websocket connection to the signal server failed."
            ),
            ApplicationError::WebSocketHandleMessageError(e) => writeln!(
                f,
                "{}: {}",
                "Couldn't handle websocket message.",
                e
            ),
            ApplicationError::RegistrationError(e) => writeln!(
                f,
                "{}: {}",
                "Registration failed.",
                e
            ),
            ApplicationError::InvalidRequest=> writeln!(
                f,
                "{}",
                "Invalid request.",
            ),
            ApplicationError::SledStore(_) => writeln!(
                f,
                "{}",
                "Something unexpected happened with the database. Please retry later."
            ),
            ApplicationError::RegistrationSuccesful => writeln!(
                f,
                "{}",
                "Registration succesful."
            ),

        }
    }
}

// convert presage errors to application errors
impl From<PresageError> for ApplicationError {
    fn from(e: PresageError) -> Self {
        match e {
            p::Error::ServiceError(p::prelude::content::ServiceError::Unauthorized) => {
                ApplicationError::UnauthorizedSignal
            }
            // p::Error::MessageSenderError(p::libsignal_service::sender::MessageSenderError {
            //     recipient: _,
            // }) => ApplicationError::NoInternet,
            p::Error::MessageSenderError(
                p::libsignal_service::sender::MessageSenderError::ServiceError(
                    p::libsignal_service::content::ServiceError::SendError { reason: e },
                ),
            ) if e.contains(FAILED_TO_LOOK_UP_ADDRESS) => ApplicationError::NoInternet,
            p::Error::MessageSenderError(e) => ApplicationError::SendFailed(e),
            _ => ApplicationError::Presage(e),
        }
    }
}
// convert presage errors to application errors
impl From<SledStoreError> for ApplicationError {
    fn from(e: SledStoreError) -> Self {
        match e {
            _ => ApplicationError::SledStore(e),
        }
    }
}
// convert websocket errors to application errors
impl From<serde_json::Error> for ApplicationError {
    fn from(e: serde_json::Error) -> Self {
        ApplicationError::WebSocketHandleMessageError(e.to_string())
    }
}

impl From<warp::Error> for ApplicationError {
    fn from(e: warp::Error) -> Self {
        ApplicationError::WebSocketHandleMessageError(e.to_string())
    }
}

impl ApplicationError {
    pub fn more_information(&self) -> String {
        match self {
            ApplicationError::NoInternet => "Please check your internet connection.".to_string(),
            ApplicationError::UnauthorizedSignal => {
                "Please delete the database and relink the device.".to_string()
            }
            ApplicationError::SendFailed(e) => format!("{:#?}", e),
            ApplicationError::ReceiveFailed(e) => format!("{:#?}", e),
            ApplicationError::Presage(e) => format!("{:#?}", e),
            ApplicationError::ManagerThreadPanic => {
                "Please restart the application with logging and report this issue.".to_string()
            }
            ApplicationError::WebSocketError => {
                "Please restart the application with logging and report this issue.".to_string()
            }
            ApplicationError::WebSocketHandleMessageError(e) => format!("{:#?}", e),
            ApplicationError::RegistrationError(e) => format!("{:#?}", e),
            ApplicationError::InvalidRequest => "Invalid request.".to_string(),
            ApplicationError::SledStore(e) => format!("{:#?}", e),
            ApplicationError::RegistrationSuccesful => "Registration succesful.".to_string(),
        }
    }

    pub fn should_report(&self) -> bool {
        match self {
            ApplicationError::NoInternet => false,
            ApplicationError::UnauthorizedSignal => false,
            ApplicationError::SendFailed(_) => false,
            ApplicationError::ReceiveFailed(_) => false,
            ApplicationError::Presage(_) => true,
            ApplicationError::ManagerThreadPanic => true,
            ApplicationError::WebSocketError => true,
            ApplicationError::WebSocketHandleMessageError(_) => true,
            ApplicationError::RegistrationError(_) => true,
            ApplicationError::InvalidRequest => false,
            ApplicationError::SledStore(_) => true,
            ApplicationError::RegistrationSuccesful => false,
        }
    }
}
