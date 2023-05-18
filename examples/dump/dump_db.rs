extern crate dirs;

fn main() {
    println!("Axolotl dump_db v.0.1.0 Starting the dump");
    let config_path = dirs::config_dir()
        .unwrap()
        .into_os_string()
        .into_string()
        .unwrap();
    let db_path = format!("{config_path}/textsecure.nanuc");

    let thedb = sled::open(db_path).unwrap();
    
    dump_registration(thedb.clone());
    dump_sessions(thedb.clone());
    dump_groups(thedb.clone());
    dump_contacts(thedb.clone());
    println!("Done dumping the database");
}

fn dump_registration(thedb: sled::Db) {
    // Iterate over all the items stored in sled to print them
    for kvr in thedb.iter() {
        if let Ok(kv) = kvr {
            let key = std::str::from_utf8(&kv.0);
            let value = std::str::from_utf8(&kv.1);
            println!("{:?} : {:?}\n", key, value);
        } else {
            println!("{:?}\n", kvr);
        }
    }
}
fn dump_sessions(thedb: sled::Db) {
    let sessions = thedb.open_tree("sessions").unwrap();
    for kvr in sessions.iter() {
        if let Ok(kv) = kvr {
            let key = std::str::from_utf8(&kv.0);
            let value = std::str::from_utf8(&kv.1);
            println!("{:?} : {:?}\n", key, value);
        } else {
            println!("{:?}\n", kvr);
        }
    }
}
fn dump_groups(thedb: sled::Db) {
    // groups are in form of a protobuf and encrypted, so we can't dump them without loading the libsignal
    let groups = thedb.open_tree("groups").unwrap();
    for kvr in groups.iter() {
        if let Ok(kv) = kvr {
            let key = std::str::from_utf8(&kv.0);
            let value = std::str::from_utf8(&kv.1);
            println!("{:?} : {:?}\n", key, value);
        } else {
            println!("{:?}\n", kvr);
        }
    }
}

fn dump_contacts(thedb: sled::Db) {
    let contacts = thedb.open_tree("contacts").unwrap();
    for kvr in contacts.iter() {
        if let Ok(kv) = kvr {
            let key = std::str::from_utf8(&kv.0);
            let value = std::str::from_utf8(&kv.1);
            println!("{:?} : {:?}\n", key, value);
        } else {
            println!("{:?}\n", kvr);
        }
    }
}

// fn dump_messages(thedb: sled::Db) {
//     let contacts = thedb.open_tree("contacts").unwrap();
//     for kvr in contacts.iter() {
//         if let Ok(kv) = kvr {
//             let uuid = Uuid::nil(); // Todo parse kv.1 as a contact and extract the uuid
//             let key = format!("threads:contact:{uuid}");
//             let mut hasher = Sha256::new();
//             hasher.update(key.as_bytes());
//         } else {
//             println!("{:?}\n", kvr);
//         }
//     }
// }

