use anyhow::Context;

/// Global Config
///
/// This struct holds the global configuration of the whisperfish app.
#[derive(serde::Serialize, serde::Deserialize, Debug)]
#[serde(rename_all = "camelCase")]
#[serde(default)]
pub struct SignalConfig {
    /// Our telephone number. This field is changed in threads and thus has to be Send/Sync but
    /// mutable at the same time.
    // XXX use the corresponding phonenumber::phonenumber type
    tel: std::sync::Mutex<String>,
    /// Our uuid. This field is changed in threads and thus has to be Send/Sync but mutable at the
    /// same time.
    // XXX use the uuid type here
    uuid: std::sync::Mutex<String>,
    /// Directory for persistent share files
    // XXX share dir is an ugly name, use another one
    // XXX we don't (de-)serialize this field as there is another instance that is accessing the
    // default path in `settings.rs`. As long as `settings.rs` is not accessing this struct, we
    // cannot set this path by a config file.
    #[serde(skip)]
    share_dir: std::path::PathBuf,
    /// Verbosity of the logging messages
    pub verbose: bool,
}

impl Default for SignalConfig {
    fn default() -> Self {
        let path =
            crate::store::default_location().expect("Could not get xdg share directory path");

        Self {
            tel: std::sync::Mutex::new(String::from("")),
            uuid: std::sync::Mutex::new(String::from("")),
            share_dir: path.to_path_buf(),
            verbose: false,
        }
    }
}

impl SignalConfig {
    pub fn read_from_file() -> Result<Self, anyhow::Error> {
        let path = dirs::config_dir()
            .context("Could not get xdg config directory path")?
            .join("harbour-whisperfish")
            .join("config.yml");

        let fd = std::fs::File::open(&path)
            .with_context(|| format!("Could not open config file: {}", &path.display()))?;
        let ret = serde_yaml::from_reader(fd)
            .with_context(|| format!("Could not read config file: {}", &path.display()))?;

        Ok(ret)
    }

    pub fn write_to_file(&self) -> Result<(), anyhow::Error> {
        let path = dirs::config_dir()
            // XXX use anyhow context here
            .expect("No config directory found")
            .join("harbour-whisperfish");

        // create config directory if it does not exist
        if !path.exists() {
            std::fs::create_dir(&path).with_context(|| {
                format!("Could not create config directory: {}", &path.display())
            })?;
        }

        // write to config file
        let path = path.join("config.yml");
        let fd = std::fs::File::create(&path)
            .with_context(|| format!("Could not open config file to write: {}", &path.display()))?;
        serde_yaml::to_writer(fd, &self)
            .with_context(|| format!("Could not write config file: {}", &path.display()))?;

        Ok(())
    }

    pub fn get_share_dir(&self) -> std::path::PathBuf {
        self.share_dir.to_owned()
    }

    // XXX should be deprecated / removed
    pub fn get_storage_dir(&self) -> std::path::PathBuf {
        self.share_dir.join("storage")
    }

    pub fn default_attachment_dir(&self) -> std::path::PathBuf {
        self.share_dir.join("storage").join("attachments")
    }

    pub fn get_identity_dir(&self) -> std::path::PathBuf {
        self.share_dir
            .join("storage")
            .join("identity")
            .join("identity_key")
    }

    pub fn get_tel_clone(&self) -> String {
        self.tel.lock().unwrap().clone()
    }

    pub fn get_uuid_clone(&self) -> String {
        self.uuid.lock().unwrap().clone()
    }

    pub fn set_tel(&self, tel: String) {
        *self.tel.lock().unwrap() = tel;
    }

    pub fn set_uuid(&self, uuid: String) {
        *self.uuid.lock().unwrap() = uuid;
    }
}
