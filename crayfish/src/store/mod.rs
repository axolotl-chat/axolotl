use std::path::{Path, PathBuf};
use std::sync::{Arc, Mutex};

use anyhow::Context;

use libsignal_service::groups_v2::InMemoryCredentialsCache;
use libsignal_service::prelude::protocol::*;

// use protocol::IdentityKeyPair;

mod protocol_store;
use protocol_store::ProtocolStore;

#[derive(Clone)]
pub struct Storage {
    // aesKey + macKey
    keys: Option<[u8; 16 + 20]>,
    pub(crate) protocol_store: Arc<tokio::sync::RwLock<ProtocolStore>>,
    credential_cache: Arc<Mutex<InMemoryCredentialsCache>>,
    path: PathBuf,
}
impl Storage {
    /// Returns the path to the storage.
    pub fn path(&self) -> &Path {
        &self.path
    }

    pub async fn open<T: AsRef<Path>>(
        storage_path: &StorageLocation<T>,
    ) -> Result<Storage, anyhow::Error> {
        let path: &Path = std::ops::Deref::deref(storage_path);
        // Axolotls keys arn't encrypted
        // let keys = Self::load_keys();

        let protocol_store = ProtocolStore::open_with_key(None, path).await?;

        Ok(Storage {
            keys: None,
            protocol_store: Arc::new(tokio::sync::RwLock::new(protocol_store)),
            credential_cache: Arc::new(Mutex::new(InMemoryCredentialsCache::default())),
            path: path.to_path_buf(),
        })
    }
    /// Asynchronously loads the base64 encoded signaling key.
    pub async fn signaling_key(&self) -> Result<[u8; 52], anyhow::Error> {
        let v = self
            .load_file(
                self.path
                    .join(".storage")
                    .join("identity")
                    .join("http_signaling_key"),
            )
            .await?;
        anyhow::ensure!(v.len() == 52, "Signaling key is 52 bytes");
        let mut out = [0u8; 52];
        out.copy_from_slice(&v);
        Ok(out)
    }
    /// Asynchronously loads the signal HTTP password from storage and decrypts it.
    pub async fn signal_password(&self) -> Result<String, anyhow::Error> {
        let contents = self
            .load_file(
                self.path
                    .join(".storage")
                    .join("identity")
                    .join("http_password"),
            )
            .await?;
        Ok(String::from_utf8(contents)?)
    }

    async fn load_file(&self, path: PathBuf) -> Result<Vec<u8>, anyhow::Error> {
        load_file(self.keys, path).await
    }
}

impl ProtocolStore {
    pub async fn open_with_key(
        keys: Option<[u8; 16 + 20]>,
        path: &Path,
    ) -> Result<Self, anyhow::Error> {
        // Identity
        let identity_path = path.join(".storage").join("identity");

        let regid = load_file(keys, identity_path.join("regid")).await?;
        let regid = String::from_utf8(regid)?;
        let regid = regid.parse()?;
        let identity_key_pair = {
            use std::convert::TryFrom;
            let buf = load_file(keys, identity_path.join("identity_key")).await?;
            let public = IdentityKey::decode(&buf[0..33])?;
            let private = PrivateKey::try_from(&buf[33..])?;
            IdentityKeyPair::new(public, private)
        };

        Ok(Self {
            identity_key_pair,
            regid,
        })
    }
}

fn load_file_sync_unencrypted(path: PathBuf) -> Result<Vec<u8>, anyhow::Error> {
    log::trace!("Opening unencrypted file {:?}", path);
    let contents = std::fs::read(&path)?;
    let count = contents.len();
    log::trace!("Read {:?}, {} bytes", path, count);
    Ok(contents)
}

fn load_file_sync_encrypted(keys: [u8; 16 + 20], path: PathBuf) -> Result<Vec<u8>, anyhow::Error> {
    // XXX This is *full* of bad practices.
    // Let's try to migrate to nacl or something alike in the future.

    log::trace!("Opening encrypted file {:?}", path);
    let mut contents = std::fs::read(&path)?;
    let count = contents.len();

    log::trace!("Read {:?}, {} bytes", path, count);
    anyhow::ensure!(count >= 16 + 32, "File smaller than cryptographic overhead");

    let (iv, contents) = contents.split_at_mut(16);
    let count = count - 16;
    let (contents, mac) = contents.split_at_mut(count - 32);

    {
        use hmac::{Hmac, Mac, NewMac};
        use sha2::Sha256;
        // Verify HMAC SHA256, 32 last bytes
        let mut verifier = Hmac::<Sha256>::new_from_slice(&keys[16..])
            .map_err(|_| anyhow::anyhow!("MAC keylength error"))?;
        verifier.update(iv);
        verifier.update(contents);
        verifier
            .verify(mac)
            .map_err(|_| anyhow::anyhow!("MAC error"))?;
    }

    use aes::Aes128;
    use block_modes::block_padding::Pkcs7;
    use block_modes::{BlockMode, Cbc};
    // Decrypt password
    let cipher = Cbc::<Aes128, Pkcs7>::new_from_slices(&keys[0..16], iv)
        .context("CBC initialization error")?;
    Ok(cipher
        .decrypt(contents)
        .context("AES CBC decryption error")?
        .to_owned())
}

fn load_file_sync(keys: Option<[u8; 16 + 20]>, path: PathBuf) -> Result<Vec<u8>, anyhow::Error> {
    match keys {
        Some(keys) => load_file_sync_encrypted(keys, path),
        None => load_file_sync_unencrypted(path),
    }
}

async fn load_file(keys: Option<[u8; 16 + 20]>, path: PathBuf) -> Result<Vec<u8>, anyhow::Error> {
    let contents = actix_threadpool::run(move || load_file_sync(keys, path)).await?;

    Ok(contents)
}

/// Location of the storage.
///
/// Path is for persistent storage.
/// Memory is for running tests or 'incognito' mode.
#[cfg_attr(not(test), allow(unused))]
pub enum StorageLocation<P> {
    Path(P),
    Memory,
}
impl<P: AsRef<Path>> std::ops::Deref for StorageLocation<P> {
    type Target = Path;
    fn deref(&self) -> &Path {
        match self {
            StorageLocation::Memory => unimplemented!(":memory: deref"),
            StorageLocation::Path(p) => p.as_ref(),
        }
    }
}

pub fn default_location() -> Result<StorageLocation<PathBuf>, anyhow::Error> {
    let data_dir = dirs::data_local_dir().context("Could not find data directory.")?;

    Ok(StorageLocation::Path(data_dir.join("harbour-whisperfish")))
}

fn write_file_sync_unencrypted(path: PathBuf, contents: &[u8]) -> Result<(), anyhow::Error> {
    log::trace!("Writing unencrypted file {:?}", path);

    use std::io::Write;
    let mut file = std::fs::File::create(&path)?;
    file.write_all(contents)?;

    Ok(())
}

fn write_file_sync_encrypted(
    keys: [u8; 16 + 20],
    path: PathBuf,
    contents: &[u8],
) -> Result<(), anyhow::Error> {
    log::trace!("Writing encrypted file {:?}", path);

    // Generate random IV
    use rand::RngCore;
    let mut iv = [0u8; 16];
    rand::thread_rng().fill_bytes(&mut iv);

    // Encrypt
    use aes::Aes128;
    use block_modes::block_padding::Pkcs7;
    use block_modes::{BlockMode, Cbc};
    let ciphertext = {
        let cipher = Cbc::<Aes128, Pkcs7>::new_from_slices(&keys[0..16], &iv)
            .context("CBC initialization error")?;
        cipher.encrypt_vec(contents)
    };

    let mac = {
        use hmac::{Hmac, Mac, NewMac};
        use sha2::Sha256;
        // Verify HMAC SHA256, 32 last bytes
        let mut mac = Hmac::<Sha256>::new_from_slice(&keys[16..])
            .map_err(|_| anyhow::anyhow!("MAC keylength error"))?;
        mac.update(&iv);
        mac.update(&ciphertext);
        mac.finalize().into_bytes()
    };

    // Write iv, ciphertext, mac
    use std::io::Write;
    let mut file = std::fs::File::create(&path)?;
    file.write_all(&iv)?;
    file.write_all(&ciphertext)?;
    file.write_all(&mac)?;

    Ok(())
}

fn write_file_sync(
    keys: Option<[u8; 16 + 20]>,
    path: PathBuf,
    contents: &[u8],
) -> Result<(), anyhow::Error> {
    match keys {
        Some(keys) => write_file_sync_encrypted(keys, path, contents),
        None => write_file_sync_unencrypted(path, contents),
    }
}

async fn write_file(
    keys: Option<[u8; 16 + 20]>,
    path: PathBuf,
    contents: Vec<u8>,
) -> Result<(), anyhow::Error> {
    actix_threadpool::run(move || write_file_sync(keys, path, &contents)).await?;
    Ok(())
}