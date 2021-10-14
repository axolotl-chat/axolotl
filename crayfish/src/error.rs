pub type Result<T> = std::result::Result<T, Error>;

#[derive(Debug)]
pub struct Error {
    pub msg: String,
}

impl Error {
    pub fn new<S: ToString>(msg: S) -> Self {
        Self {
            msg: msg.to_string(),
        }
    }
    pub fn is<S: ToString, T>(msg: S) -> Result<T> {
        Err(Self::new(msg))
    }
}

impl<E: ToString> From<E> for Error {
    fn from(error: E) -> Self {
        Self {
            msg: error.to_string(),
        }
    }
}
