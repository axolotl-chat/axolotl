//! Quirks for the on-disk structures of textsecure.
//!
//! Textsecure uses the same on-disk protobuf format as libsignal-protocol, however, some of the
//! byte-array fields have a quirky behaviour. This module provides methods to add and remove those
//! quirks.
//!
//! This module maps on https://gitlab.com/rubdos/whisperfish/-/issues/74.

use libsignal_service::prelude::protocol::SignalProtocolError;
use prost::Message;

include!(concat!(env!("OUT_DIR"), "/textsecure.rs"));

pub const DJB_TYPE: u8 = 0x05;

/// Removes quirks to the session data format that are apparent in Whisperfish 0.5
pub fn session_from_0_5(input: &[u8]) -> Result<Vec<u8>, SignalProtocolError> {
    let mut obj = RecordStructure::decode(input)?;

    // begin unquirking
    obj.current_session
        .as_mut()
        .map(unquirk_session_structure)
        .transpose()?;
    for session in &mut obj.previous_sessions {
        unquirk_session_structure(session)?;
    }
    // end unquirking

    Ok(obj.encode_to_vec())
}

/// Adds quirks to the session data format that are apparent in Whisperfish 0.5
pub fn session_to_0_5(input: &[u8]) -> Result<Vec<u8>, SignalProtocolError> {
    let mut obj = RecordStructure::decode(input)?;

    // begin quirking
    obj.current_session
        .as_mut()
        .map(quirk_session_structure)
        .transpose()?;
    for session in &mut obj.previous_sessions {
        quirk_session_structure(session)?;
    }
    // end quirking

    Ok(obj.encode_to_vec())
}

/// Removes quirks to the pre key data format that are apparent in Whisperfish 0.5
pub fn pre_key_from_0_5(input: &[u8]) -> Result<Vec<u8>, SignalProtocolError> {
    let mut obj = PreKeyRecordStructure::decode(input)?;

    // begin quirking
    unquirk_identity(&mut obj.public_key)?;
    // end quirking

    Ok(obj.encode_to_vec())
}

/// Adds quirks to the pre key data format that are apparent in Whisperfish 0.5
pub fn pre_key_to_0_5(input: &[u8]) -> Result<Vec<u8>, SignalProtocolError> {
    let mut obj = PreKeyRecordStructure::decode(input)?;

    // begin quirking
    quirk_identity(&mut obj.public_key)?;
    // end quirking

    Ok(obj.encode_to_vec())
}

/// Removes quirks to the signed pre key data format that are apparent in Whisperfish 0.5
pub fn signed_pre_key_from_0_5(input: &[u8]) -> Result<Vec<u8>, SignalProtocolError> {
    let mut obj = SignedPreKeyRecordStructure::decode(input)?;

    // begin quirking
    unquirk_identity(&mut obj.public_key)?;
    // end quirking

    Ok(obj.encode_to_vec())
}

/// Adds quirks to the signed pre key data format that are apparent in Whisperfish 0.5
pub fn signed_pre_key_to_0_5(input: &[u8]) -> Result<Vec<u8>, SignalProtocolError> {
    let mut obj = SignedPreKeyRecordStructure::decode(input)?;

    // begin quirking
    quirk_identity(&mut obj.public_key)?;
    // end quirking

    Ok(obj.encode_to_vec())
}

fn quirky_keys_mut(sess: &mut SessionStructure) -> impl Iterator<Item = &mut Vec<u8>> {
    let chains = std::iter::once(sess.sender_chain.as_mut())
        .flatten() // filter out Option<_>
        .chain(sess.receiver_chains.iter_mut())
        .map(|chain| &mut chain.sender_ratchet_key);

    vec![
        sess.local_identity_public.as_mut(),
        sess.remote_identity_public.as_mut(),
        // sess.alice_base_key.as_mut(), // Alice base key, for some reason, is not quirky
    ]
    .into_iter()
    .chain(
        sess.pending_pre_key
            .as_mut()
            .into_iter()
            .map(|ppk| ppk.base_key.as_mut()),
    )
    .chain(chains)
}

fn quirk_session_structure(sess: &mut SessionStructure) -> Result<(), SignalProtocolError> {
    for identity in quirky_keys_mut(sess) {
        quirk_identity(identity)?;
    }

    Ok(())
}

fn unquirk_session_structure(sess: &mut SessionStructure) -> Result<(), SignalProtocolError> {
    for identity in quirky_keys_mut(sess) {
        unquirk_identity(identity)?;
    }

    Ok(())
}

fn quirk_identity(id: &mut Vec<u8>) -> Result<(), SignalProtocolError> {
    if id.len() == 32 {
        log::warn!("Not quirking input key of 32 bytes!");
        Ok(())
    } else if id.len() == 32 + 1 {
        let removed = id.remove(0);
        if removed != DJB_TYPE {
            log::error!("Unknown input key type {}, not quirking.", removed);
            Err(SignalProtocolError::InternalError("Unknown key type"))
        } else {
            Ok(())
        }
    } else {
        log::error!("Invalid input key of length {}", id.len());
        Err(SignalProtocolError::InternalError(
            "Invalid identity key length",
        ))
    }
}

fn unquirk_identity(id: &mut Vec<u8>) -> Result<(), SignalProtocolError> {
    if id.len() == 33 {
        log::warn!(
            "Not unquirking input key of 33 bytes! Its tarts with {}.",
            id[0]
        );
        Ok(())
    } else if id.len() == 32 {
        id.insert(0, DJB_TYPE);
        Ok(())
    } else {
        log::error!("Invalid input key of length {}, cannot unquirk", id.len());
        Err(SignalProtocolError::InternalError(
            "Invalid identity key length",
        ))
    }
}
