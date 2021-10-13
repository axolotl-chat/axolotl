use std::io;
use std::path::Path;

use libsignal_service::prelude::protocol::{self, Context};
use protocol::IdentityKeyPair;
use protocol::SignalProtocolError;

mod quirk;

use super::*;

pub struct ProtocolStore {
    pub(crate) identity_key_pair: IdentityKeyPair,
    pub(crate) regid: u32,
}

fn convert_io_error(e: io::Error) -> SignalProtocolError {
    // XXX can probably be better, but currently this is only used in session_delete and
    // identity_delete
    SignalProtocolError::SessionNotFound(e.to_string())
}

fn addr_to_path_component<'a>(addr: &'a (impl AsRef<[u8]> + ?Sized + 'a)) -> &'a str {
    let addr: &'a [u8] = addr.as_ref();
    let addr = if addr[0] == b'+' { &addr[1..] } else { addr };
    std::str::from_utf8(addr).expect("address in valid UTF8")
}

impl ProtocolStore {
    pub async fn store_with_key(
        keys: Option<[u8; 16 + 20]>,
        path: &Path,
        regid: u32,
        identity_key_pair: IdentityKeyPair,
    ) -> Result<Self, anyhow::Error> {
        // Identity
        let identity_path = path.join("storage").join("identity");

        // XXX move to quirk
        let mut identity_key = Vec::new();
        let public = identity_key_pair.public_key().serialize();
        assert_eq!(public.len(), 32 + 1);
        assert_eq!(public[0], quirk::DJB_TYPE);
        identity_key.extend(&public[1..]);

        let private = identity_key_pair.private_key().serialize();
        assert_eq!(private.len(), 32);
        identity_key.extend(private);

        write_file(
            keys,
            identity_path.join("regid"),
            format!("{}", regid).into_bytes(),
        )
        .await?;
        write_file(keys, identity_path.join("identity_key"), identity_key).await?;

        Ok(Self {
            identity_key_pair,
            regid,
        })
    }

}

impl Storage {
    fn session_path(&self, addr: &ProtocolAddress) -> PathBuf {
        let recipient_id = addr_to_path_component(addr.name());

        self.path.join("storage").join("sessions").join(format!(
            "{}_{}",
            recipient_id,
            addr.device_id()
        ))
    }

    fn identity_path(&self, addr: &ProtocolAddress) -> PathBuf {
        let recipient_id = addr_to_path_component(addr.name());

        self.path
            .join("storage")
            .join("identity")
            .join(format!("remote_{}", recipient_id,))
    }

    fn prekey_path(&self, id: u32) -> PathBuf {
        self.path
            .join("storage")
            .join("prekeys")
            .join(format!("{:09}", id))
    }

    fn signed_prekey_path(&self, id: u32) -> PathBuf {
        self.path
            .join("storage")
            .join("signed_prekeys")
            .join(format!("{:09}", id))
    }

    /// Returns a tuple of the next free signed pre-key ID and the next free pre-key ID
    pub async fn next_pre_key_ids(&self) -> (u32, u32) {
        let _lock = self.protocol_store.read().await;

        let mut pre_key_ids: Vec<u32> =
            std::fs::read_dir(self.path.join("storage").join("prekeys"))
                .expect("initialized storage")
                .filter_map(|entry| {
                    let entry = entry.expect("directory listing");
                    if !entry.path().is_file() {
                        log::warn!("Non-file session entry: {:?}. Skipping", entry);
                        return None;
                    }

                    // XXX: *maybe* Signal could become a cross-platform desktop app.
                    use std::os::unix::ffi::OsStrExt;
                    let name = entry.file_name();
                    let name = name.as_os_str().as_bytes();

                    log::trace!("parsing {:?}", entry);
                    let id = std::str::from_utf8(name).ok()?;
                    id.parse().ok()
                })
                .collect();
        pre_key_ids.sort_unstable();

        let mut signed_pre_key_ids: Vec<u32> =
            std::fs::read_dir(self.path.join("storage").join("signed_prekeys"))
                .expect("initialized storage")
                .filter_map(|entry| {
                    let entry = entry.expect("directory listing");
                    if !entry.path().is_file() {
                        log::warn!("Non-file session entry: {:?}. Skipping", entry);
                        return None;
                    }

                    // XXX: *maybe* Signal could become a cross-platform desktop app.
                    use std::os::unix::ffi::OsStrExt;
                    let name = entry.file_name();
                    let name = name.as_os_str().as_bytes();

                    log::trace!("parsing {:?}", entry);
                    let id = std::str::from_utf8(name).ok()?;
                    id.parse().ok()
                })
                .collect();
        signed_pre_key_ids.sort_unstable();

        let next_pre_key_id = if pre_key_ids.is_empty() {
            0
        } else {
            pre_key_ids[pre_key_ids.len() - 1] + 1
        };
        let next_signed_pre_key_id = if signed_pre_key_ids.is_empty() {
            0
        } else {
            signed_pre_key_ids[signed_pre_key_ids.len() - 1] + 1
        };
        (next_signed_pre_key_id, next_pre_key_id)
    }

    pub async fn delete_identity(&self, addr: &ProtocolAddress) -> Result<(), SignalProtocolError> {
        let _lock = self.protocol_store.write().await;

        let path = self.identity_path(addr);
        std::fs::remove_file(path).map_err(convert_io_error)?;
        Ok(())
    }
}

#[async_trait::async_trait(?Send)]
impl protocol::IdentityKeyStore for Storage {
    async fn get_identity_key_pair(
        &self,
        _: Context,
    ) -> Result<IdentityKeyPair, SignalProtocolError> {
        log::trace!("identity_key_pair");
        let protocol_store = self.protocol_store.read().await;
        Ok(protocol_store.identity_key_pair)
    }

    async fn get_local_registration_id(&self, _: Context) -> Result<u32, SignalProtocolError> {
        Ok(self.protocol_store.read().await.regid)
    }

    async fn is_trusted_identity(
        &self,
        addr: &ProtocolAddress,
        key: &IdentityKey,
        // XXX
        _direction: Direction,
        _ctx: Context,
    ) -> Result<bool, SignalProtocolError> {
        let _lock = self.protocol_store.read().await;

        if let Some(trusted_key) = self.read_identity_key_file(addr).await? {
            Ok(trusted_key == *key)
        } else {
            // Trust on first use
            Ok(true)
        }
    }

    /// Should return true when the older key, if present, is different from the new one.
    /// False otherwise.
    async fn save_identity(
        &mut self,
        addr: &ProtocolAddress,
        key: &IdentityKey,
        _: Context,
    ) -> Result<bool, SignalProtocolError> {
        let _lock = self.protocol_store.write().await;

        // Save return value
        let mut ret = false;

        // Get old key if present and compare to the new one. If they are the same, we set `ret` to
        // true.
        if let Some(key_old) = self.read_identity_key_file(addr).await? {
            ret = key_old == *key;
        };

        // Save new key only if ret is false. If it is `true` the old and the new key are identical
        // and saving is not necessary
        if !ret {
            // Write key
            let path = self.identity_path(addr);
            write_file(self.keys, path, key.serialize().into())
                .await
                .expect("save identity key");
        }

        Ok(ret)
    }

    async fn get_identity(
        &self,
        addr: &ProtocolAddress,
        _: Context,
    ) -> Result<Option<IdentityKey>, SignalProtocolError> {
        let _lock = self.protocol_store.read().await;

        self.read_identity_key_file(addr).await
    }
}

#[async_trait::async_trait(?Send)]
impl protocol::PreKeyStore for Storage {
    async fn get_pre_key(&self, id: u32, _: Context) -> Result<PreKeyRecord, SignalProtocolError> {
        log::trace!("Loading prekey {}", id);
        let _lock = self.protocol_store.read().await;

        let path = self.prekey_path(id);
        let contents = if let Ok(x) = load_file(self.keys, path).await {
            x
        } else {
            return Err(SignalProtocolError::InvalidPreKeyId);
        };
        let contents = quirk::pre_key_from_0_5(&contents).unwrap();
        Ok(PreKeyRecord::deserialize(&contents)?)
    }

    async fn save_pre_key(
        &mut self,
        id: u32,
        body: &PreKeyRecord,
        _: Context,
    ) -> Result<(), SignalProtocolError> {
        log::trace!("Storing prekey {}", id);
        let _lock = self.protocol_store.write().await;

        let path = self.prekey_path(id);
        let contents = quirk::pre_key_to_0_5(&body.serialize()?).unwrap();
        write_file(self.keys, path, contents)
            .await
            .expect("written file");
        Ok(())
    }

    async fn remove_pre_key(&mut self, id: u32, _: Context) -> Result<(), SignalProtocolError> {
        log::trace!("Removing prekey {}", id);
        let _lock = self.protocol_store.write().await;

        let path = self.prekey_path(id);
        std::fs::remove_file(path).map_err(|_| SignalProtocolError::InvalidPreKeyId)?;
        Ok(())
    }
}

impl Storage {
    // XXX Rewrite in terms of get_pre_key
    #[allow(dead_code)]
    async fn contains_pre_key(&self, id: u32) -> bool {
        log::trace!("Checking for prekey {}", id);
        let _lock = self.protocol_store.read().await;

        self.prekey_path(id).is_file()
    }
}

#[async_trait::async_trait(?Send)]
impl protocol::SessionStore for Storage {
    async fn load_session(
        &self,
        addr: &ProtocolAddress,
        _: Context,
    ) -> Result<Option<SessionRecord>, SignalProtocolError> {
        let path = self.session_path(addr);

        log::trace!("Loading session for {:?} from {:?}", addr, path);
        let _lock = self.protocol_store.read().await;

        let buf = if let Ok(buf) = load_file(self.keys, path).await {
            quirk::session_from_0_5(&buf)?
        } else {
            return Ok(None);
        };

        Ok(Some(SessionRecord::deserialize(&buf)?))
    }

    async fn store_session(
        &mut self,
        addr: &ProtocolAddress,
        session: &protocol::SessionRecord,
        _: Context,
    ) -> Result<(), SignalProtocolError> {
        let path = self.session_path(addr);

        log::trace!("Storing session for {:?} at {:?}", addr, path);
        let _lock = self.protocol_store.write().await;

        let quirked = quirk::session_to_0_5(&session.serialize()?)?;
        write_file(self.keys, path, quirked).await.unwrap();
        Ok(())
    }
}

impl Storage {
    #[allow(dead_code)]
    async fn contains_session(
        &self,
        addr: &ProtocolAddress,
        _: Context,
    ) -> Result<bool, SignalProtocolError> {
        let _lock = self.protocol_store.read().await;

        let path = self.session_path(addr);
        Ok(path.is_file())
    }
}

#[cfg_attr(feature = "unsend-futures", async_trait::async_trait(?Send))]
#[cfg_attr(not(feature = "unsend-futures"), async_trait::async_trait)]
impl protocol::SessionStoreExt for Storage {
    async fn get_sub_device_sessions(&self, addr: &str) -> Result<Vec<u32>, SignalProtocolError> {
        log::trace!("Looking for sub_device sessions for {}", addr);
        let _lock = self.protocol_store.read().await;

        let addr = addr_to_path_component(addr).as_bytes();

        let session_dir = self.path.join("storage").join("sessions");

        let ids = std::fs::read_dir(session_dir)
            .expect("initialized storage")
            .filter_map(|entry| {
                let entry = entry.expect("directory listing");
                if !entry.path().is_file() {
                    log::warn!("Non-file session entry: {:?}. Skipping", entry);
                    return None;
                }

                // XXX: *maybe* Signal could become a cross-platform desktop app.
                use std::os::unix::ffi::OsStrExt;
                let name = entry.file_name();
                let name = name.as_os_str().as_bytes();

                if name.len() < addr.len() + 2 {
                    return None;
                }

                if &name[..addr.len()] == addr {
                    if name[addr.len()] != b'_' {
                        log::warn!("Weird session directory entry: {:?}. Skipping", entry);
                        return None;
                    }
                    // skip underscore
                    let id = std::str::from_utf8(&name[(addr.len() + 1)..]).ok()?;
                    id.parse().ok()
                } else {
                    None
                }
            })
            .filter(|id| *id != libsignal_service::push_service::DEFAULT_DEVICE_ID)
            .collect();

        Ok(ids)
    }

    async fn delete_session(&self, addr: &ProtocolAddress) -> Result<(), SignalProtocolError> {
        let _lock = self.protocol_store.write().await;

        let path = self.session_path(addr);
        std::fs::remove_file(path).map_err(|e| {
            log::debug!(
                "Could not delete session {}, assuming non-existing: {}",
                addr.to_string(),
                e
            );
            SignalProtocolError::SessionNotFound(addr.to_string())
        })?;
        Ok(())
    }

    async fn delete_all_sessions(&self, addr: &str) -> Result<usize, SignalProtocolError> {
        log::warn!("Deleting all sessions for {}", addr);
        let _lock = self.protocol_store.write().await;

        let addr = addr_to_path_component(addr).as_bytes();

        let session_dir = self.path.join("storage").join("sessions");

        let entries = std::fs::read_dir(session_dir)
            .expect("initialized storage")
            .filter_map(|entry| {
                let entry = entry.expect("directory listing");
                if !entry.path().is_file() {
                    log::warn!("Non-file session entry: {:?}. Skipping", entry);
                    return None;
                }

                // XXX: *maybe* Signal could become a cross-platform desktop app.
                use std::os::unix::ffi::OsStrExt;
                let name = entry.file_name();
                let name = name.as_os_str().as_bytes();

                log::trace!("parsing {:?}", entry);

                if name.len() < addr.len() + 2 {
                    log::trace!("filename {:?} not long enough", entry);
                    return None;
                }

                if &name[..addr.len()] == addr {
                    if name[addr.len()] != b'_' {
                        log::warn!("Weird session directory entry: {:?}. Skipping", entry);
                        return None;
                    }
                    // skip underscore
                    let id = std::str::from_utf8(&name[(addr.len() + 1)..]).ok()?;
                    let _: u32 = id.parse().ok()?;
                    Some(entry.path())
                } else {
                    log::trace!("filename {:?} without prefix match", entry);
                    None
                }
            });

        let mut count = 0;
        for entry in entries {
            std::fs::remove_file(entry).map_err(convert_io_error)?;
            count += 1;
        }

        Ok(count)
    }
}

#[async_trait::async_trait(?Send)]
impl protocol::SignedPreKeyStore for Storage {
    async fn get_signed_pre_key(
        &self,
        id: u32,
        _: Context,
    ) -> Result<SignedPreKeyRecord, SignalProtocolError> {
        log::trace!("Loading signed prekey {}", id);
        let _lock = self.protocol_store.read().await;

        let path = self.signed_prekey_path(id);

        let contents = if let Ok(x) = load_file(self.keys, path).await {
            x
        } else {
            return Err(SignalProtocolError::InvalidSignedPreKeyId);
        };
        let contents = quirk::signed_pre_key_from_0_5(&contents).unwrap();

        Ok(SignedPreKeyRecord::deserialize(&contents)?)
    }

    async fn save_signed_pre_key(
        &mut self,
        id: u32,
        body: &SignedPreKeyRecord,
        _: Context,
    ) -> Result<(), SignalProtocolError> {
        log::trace!("Storing prekey {}", id);
        let _lock = self.protocol_store.write().await;

        let path = self.signed_prekey_path(id);
        let contents = quirk::signed_pre_key_to_0_5(&body.serialize()?).unwrap();
        write_file(self.keys, path, contents)
            .await
            .expect("written file");
        Ok(())
    }
}

impl Storage {
    #[allow(dead_code)]
    async fn remove_signed_pre_key(&self, id: u32) -> Result<(), SignalProtocolError> {
        log::trace!("Removing signed prekey {}", id);
        let _lock = self.protocol_store.write().await;

        let path = self.signed_prekey_path(id);
        std::fs::remove_file(path).map_err(|_| SignalProtocolError::InvalidPreKeyId)?;
        Ok(())
    }

    // XXX rewrite in terms of get_signed_pre_key
    #[allow(dead_code)]
    async fn contains_signed_pre_key(&self, id: u32) -> bool {
        log::trace!("Checking for signed prekey {}", id);
        let _lock = self.protocol_store.read().await;

        self.signed_prekey_path(id).is_file()
    }

    async fn read_identity_key_file(
        &self,
        addr: &ProtocolAddress,
    ) -> Result<Option<IdentityKey>, SignalProtocolError> {
        let path = self.identity_path(addr);
        if path.is_file() {
            let buf = load_file(self.keys, path).await.expect("read identity key");
            match buf.len() {
                // Old format
                32 => Ok(Some(
                    protocol::PublicKey::from_djb_public_key_bytes(&buf)?.into(),
                )),
                // New format
                33 => Ok(Some(IdentityKey::decode(&buf)?)),
                _ => Err(SignalProtocolError::InvalidArgument(format!(
                    "Identity key has length {}, expected 32 or 33",
                    buf.len()
                ))),
            }
        } else {
            Ok(None)
        }
    }
}

#[cfg(test)]
mod tests {
    use libsignal_service::prelude::protocol::*;
    use rstest::rstest;

    async fn create_example_storage(
        storage_password: Option<&str>,
    ) -> Result<(super::Storage, super::StorageLocation<tempdir::TempDir>), anyhow::Error> {
        use rand::distributions::Alphanumeric;
        use rand::{Rng, RngCore};

        env_logger::try_init().ok();

        let location = super::temp();
        let rng = rand::thread_rng();

        // Signaling password for REST API
        let password: String = rng.sample_iter(&Alphanumeric).take(24).collect();

        // Signaling key that decrypts the incoming Signal messages
        let mut rng = rand::thread_rng();
        let mut signaling_key = [0u8; 52];
        rng.fill_bytes(&mut signaling_key);
        let signaling_key = signaling_key;

        // Registration ID
        let regid = 12345;

        let storage = super::Storage::new(
            &location,
            storage_password.as_deref(),
            regid,
            &password,
            signaling_key,
        )
        .await?;

        Ok((storage, location))
    }

    fn create_random_protocol_address() -> ProtocolAddress {
        use rand::Rng;
        let mut rng = rand::thread_rng();

        let user_id = uuid::Uuid::new_v4();
        let device_id = rng.gen_range(2, 20);

        ProtocolAddress::new(user_id.to_string(), device_id)
    }

    fn create_random_identity_key() -> IdentityKey {
        let mut rng = rand::thread_rng();

        let key_pair = IdentityKeyPair::generate(&mut rng);

        *key_pair.identity_key()
    }

    fn create_random_prekey() -> PreKeyRecord {
        use rand::Rng;
        let mut rng = rand::thread_rng();

        let key_pair = KeyPair::generate(&mut rng);
        let id: u32 = rng.gen();

        PreKeyRecord::new(id, &key_pair)
    }

    fn create_random_signed_prekey() -> SignedPreKeyRecord {
        use rand::Rng;
        let mut rng = rand::thread_rng();

        let key_pair = KeyPair::generate(&mut rng);
        let id: u32 = rng.gen();
        let timestamp: u64 = rng.gen();
        let signature = vec![0; 3];

        SignedPreKeyRecord::new(id, timestamp, &key_pair, &signature)
    }

    /// XXX Right now, this functions seems a bit unnecessary, but we will change the creation of a
    /// storage and it might be necessary to check the own identity_key_pair in the protocol store.
    #[rstest(password, case(Some("some password")), case(None))]
    #[actix_rt::test]
    async fn own_identity_key_pair(password: Option<&str>) {
        env_logger::try_init().ok();

        // create a new storage
        let (storage, _tempdir) = create_example_storage(password).await.unwrap();

        // Copy the identity key pair
        let id_key1 = storage.protocol_store.read().await.identity_key_pair;

        // Get access to the protocol store
        // XXX IdentityKeyPair does not implement the std::fmt::Debug trait *arg*
        //assert_eq!(id_key1.unwrap(), store.get_identity_key_pair(None).await.unwrap());
        assert_eq!(
            id_key1.serialize(),
            storage
                .get_identity_key_pair(None)
                .await
                .unwrap()
                .serialize()
        );
    }

    /// XXX Right now, this functions seems a bit unnecessary, but we will change the creation of a
    /// storage and it might be necessary to check the regid in the protocol store.
    #[rstest(password, case(Some("some password")), case(None))]
    #[actix_rt::test]
    async fn own_regid(password: Option<&str>) {
        env_logger::try_init().ok();

        // create a new storage
        let (storage, _tempdir) = create_example_storage(password).await.unwrap();

        // Copy the regid
        let regid_1 = storage.protocol_store.read().await.regid;

        // Get access to the protocol store
        assert_eq!(
            regid_1,
            storage.get_local_registration_id(None).await.unwrap()
        );
    }

    #[rstest(password, case(Some("some password")), case(None))]
    #[actix_rt::test]
    async fn save_retrieve_identity_key(password: Option<&str>) {
        env_logger::try_init().ok();

        // Create a new storage
        let (mut storage, _tempdir) = create_example_storage(password).await.unwrap();

        // We need two identity keys and two addresses
        let addr1 = create_random_protocol_address();
        let addr2 = create_random_protocol_address();
        let key1 = create_random_identity_key();
        let key2 = create_random_identity_key();

        // In the beginning, the storage should be emtpy and return an error
        // XXX Doesn't implement equality *arg*
        assert_eq!(storage.get_identity(&addr1, None).await.unwrap(), None);
        assert_eq!(storage.get_identity(&addr2, None).await.unwrap(), None);

        // We store both keys and should get false because there wasn't a key with that address
        // yet
        assert!(!storage.save_identity(&addr1, &key1, None).await.unwrap());
        assert!(!storage.save_identity(&addr2, &key2, None).await.unwrap());

        // Now, we should get both keys
        assert_eq!(
            storage.get_identity(&addr1, None).await.unwrap(),
            Some(key1)
        );
        assert_eq!(
            storage.get_identity(&addr2, None).await.unwrap(),
            Some(key2)
        );

        // After removing key2, it shouldn't be there
        storage.delete_identity(&addr2).await.unwrap();
        // XXX Doesn't implement equality *arg*
        assert_eq!(storage.get_identity(&addr2, None).await.unwrap(), None);

        // We can now overwrite key1 with key1 and should get true returned
        assert!(storage.save_identity(&addr1, &key1, None).await.unwrap());

        // We can now overwrite key1 with key2 and should get false returned
        assert!(!storage.save_identity(&addr1, &key2, None).await.unwrap());
    }

    // Direction does not matter yet
    #[rstest(password, case(Some("some password")), case(None))]
    #[actix_rt::test]
    async fn is_trusted_identity(password: Option<&str>) {
        env_logger::try_init().ok();

        // Create a new storage
        let (mut storage, _tempdir) = create_example_storage(password).await.unwrap();

        // We need two identity keys and two addresses
        let addr1 = create_random_protocol_address();
        let key1 = create_random_identity_key();
        let key2 = create_random_identity_key();

        // Test trust on first use
        assert!(storage
            .is_trusted_identity(&addr1, &key1, Direction::Receiving, None)
            .await
            .unwrap());

        // Test inserted key
        storage.save_identity(&addr1, &key1, None).await.unwrap();
        assert!(storage
            .is_trusted_identity(&addr1, &key1, Direction::Receiving, None)
            .await
            .unwrap());

        // Test wrong key
        assert!(!storage
            .is_trusted_identity(&addr1, &key2, Direction::Receiving, None)
            .await
            .unwrap());
    }

    #[rstest(password, case(Some("some password")), case(None))]
    #[actix_rt::test]
    async fn save_retrieve_prekey(password: Option<&str>) {
        env_logger::try_init().ok();

        // Create a new storage
        let (mut storage, _tempdir) = create_example_storage(password).await.unwrap();

        // We need two identity keys and two addresses
        let id1 = 0u32;
        let id2 = 1u32;
        let key1 = create_random_prekey();
        let key2 = create_random_prekey();

        // In the beginning, the storage should be emtpy and return an error
        // XXX Doesn't implement equality *arg*
        assert_eq!(
            storage
                .get_pre_key(id1, None)
                .await
                .unwrap_err()
                .to_string(),
            SignalProtocolError::InvalidPreKeyId.to_string()
        );

        // Storing both keys and testing retrieval
        storage.save_pre_key(id1, &key1, None).await.unwrap();
        storage.save_pre_key(id2, &key2, None).await.unwrap();

        // Now, we should get both keys
        assert_eq!(
            storage
                .get_pre_key(id1, None)
                .await
                .unwrap()
                .serialize()
                .unwrap(),
            key1.serialize().unwrap()
        );
        assert_eq!(
            storage
                .get_pre_key(id2, None)
                .await
                .unwrap()
                .serialize()
                .unwrap(),
            key2.serialize().unwrap()
        );

        // After removing key2, it shouldn't be there
        storage.remove_pre_key(id2, None).await.unwrap();
        // XXX Doesn't implement equality *arg*
        assert_eq!(
            storage
                .get_pre_key(id2, None)
                .await
                .unwrap_err()
                .to_string(),
            SignalProtocolError::InvalidPreKeyId.to_string()
        );

        // Let's check whether we can overwrite a key
        storage.save_pre_key(id1, &key2, None).await.unwrap();
    }

    #[rstest(password, case(Some("some password")), case(None))]
    #[actix_rt::test]
    async fn save_retrieve_signed_prekey(password: Option<&str>) {
        env_logger::try_init().ok();

        // Create a new storage
        let (mut storage, _tempdir) = create_example_storage(password).await.unwrap();

        // We need two identity keys and two addresses
        let id1 = 0u32;
        let id2 = 1u32;
        let key1 = create_random_signed_prekey();
        let key2 = create_random_signed_prekey();

        // In the beginning, the storage should be emtpy and return an error
        // XXX Doesn't implement equality *arg*
        assert_eq!(
            storage
                .get_signed_pre_key(id1, None)
                .await
                .unwrap_err()
                .to_string(),
            SignalProtocolError::InvalidSignedPreKeyId.to_string()
        );

        // Storing both keys and testing retrieval
        storage.save_signed_pre_key(id1, &key1, None).await.unwrap();
        storage.save_signed_pre_key(id2, &key2, None).await.unwrap();

        // Now, we should get both keys
        assert_eq!(
            storage
                .get_signed_pre_key(id1, None)
                .await
                .unwrap()
                .serialize()
                .unwrap(),
            key1.serialize().unwrap()
        );
        assert_eq!(
            storage
                .get_signed_pre_key(id2, None)
                .await
                .unwrap()
                .serialize()
                .unwrap(),
            key2.serialize().unwrap()
        );

        // Let's check whether we can overwrite a key
        storage.save_signed_pre_key(id1, &key2, None).await.unwrap();
    }

    #[rstest(password, case(Some("some password")), case(None))]
    #[actix_rt::test]
    async fn save_retrieve_session(password: Option<&str>) {
        env_logger::try_init().ok();

        // Create a new storage
        let (mut storage, _tempdir) = create_example_storage(password).await.unwrap();

        // Collection of some addresses and sessions
        let addr1 = create_random_protocol_address();
        let addr2 = create_random_protocol_address();
        let addr3 = create_random_protocol_address();
        let addr4 = ProtocolAddress::new(addr3.name().to_string(), addr3.device_id() + 1);
        let session1 = SessionRecord::new_fresh();
        let session2 = SessionRecord::new_fresh();
        let session3 = SessionRecord::new_fresh();
        let session4 = SessionRecord::new_fresh();

        // In the beginning, the storage should be emtpy and return an error
        assert!(storage.load_session(&addr1, None).await.unwrap().is_none());
        assert!(storage.load_session(&addr2, None).await.unwrap().is_none());

        // Store all four sessions: three different names, one name with two different device ids.
        storage
            .store_session(&addr1, &session1, None)
            .await
            .unwrap();
        storage
            .store_session(&addr2, &session2, None)
            .await
            .unwrap();
        storage
            .store_session(&addr3, &session3, None)
            .await
            .unwrap();
        storage
            .store_session(&addr4, &session4, None)
            .await
            .unwrap();

        // Now, we should get the sessions to the first two addresses
        assert_eq!(
            storage
                .load_session(&addr1, None)
                .await
                .unwrap()
                .unwrap()
                .serialize()
                .unwrap(),
            session1.serialize().unwrap()
        );
        assert_eq!(
            storage
                .load_session(&addr2, None)
                .await
                .unwrap()
                .unwrap()
                .serialize()
                .unwrap(),
            session2.serialize().unwrap()
        );

        // Let's check whether we can overwrite a key
        storage
            .store_session(&addr1, &session2, None)
            .await
            .expect("Overwrite session");

        // Get all device ids for the same address
        let mut ids = storage.get_sub_device_sessions(addr3.name()).await.unwrap();
        ids.sort_unstable();
        assert_eq!(ids[0], std::cmp::min(addr3.device_id(), addr4.device_id()));
        assert_eq!(ids[1], std::cmp::max(addr3.device_id(), addr4.device_id()));

        // If we call delete all sessions, all sessions of one person/address should be removed
        assert_eq!(storage.delete_all_sessions(addr3.name()).await.unwrap(), 2);
        assert!(storage.load_session(&addr3, None).await.unwrap().is_none());
        assert!(storage.load_session(&addr4, None).await.unwrap().is_none());

        // If we delete the first two sessions, they shouldn't be in the store anymore
        SessionStoreExt::delete_session(&storage, &addr1)
            .await
            .unwrap();
        SessionStoreExt::delete_session(&storage, &addr2)
            .await
            .unwrap();
        assert!(storage.load_session(&addr1, None).await.unwrap().is_none());
        assert!(storage.load_session(&addr2, None).await.unwrap().is_none());
    }

    #[rstest(password, case(Some("some password")), case(None))]
    #[actix_rt::test]
    async fn get_next_pre_key_ids(password: Option<&str>) {
        env_logger::try_init().ok();

        // Create a new storage
        let (mut storage, _tempdir) = create_example_storage(password).await.unwrap();

        // Create two pre keys and one signed pre key
        let key1 = create_random_prekey();
        let key2 = create_random_prekey();
        let key3 = create_random_signed_prekey();

        // In the beginning zero should be returned
        assert_eq!(storage.next_pre_key_ids().await, (0, 0));

        // Now, we add our keys
        storage.save_pre_key(0, &key1, None).await.unwrap();
        storage.save_pre_key(1, &key2, None).await.unwrap();
        storage.save_signed_pre_key(0, &key3, None).await.unwrap();

        // Adapt to keys in the storage
        assert_eq!(storage.next_pre_key_ids().await, (1, 2));
    }
}
