pub(crate) use libsignal_service::prelude::Uuid;
use serde_yaml::from_reader;
use serde::Deserialize;
use std::fs::File;

const CONFIG_PATH : &str = "path/to/config";

#[allow(dead_code)]
pub fn migrate_config()-> Result<(), String> {
    let mut reader = YamlConfigReader{ path: CONFIG_PATH };
    migrate(&mut reader)
}

fn migrate(reader: &mut impl ConfigReader) -> Result<(), String> {
    let _config = reader.read_config()?;
    // todo do stuff
    Ok(())
}

#[derive(Deserialize, PartialEq, Debug)]
struct Config {
    pub name: String,
    pub tel: String,
    pub uuid: Uuid,
    #[serde(rename = "profileKey")]
    pub profile_key: Vec<u8>,
    #[serde(rename = "profileKeyCredential")]
    pub profile_key_credential: Vec<u8>,
    pub certificate: Vec<u8>
}

trait ConfigReader {
    fn read_config(&mut self) -> Result<Box<Config>,String>;
}

pub struct YamlConfigReader {
    pub path: &'static str
}

impl ConfigReader for YamlConfigReader {
    fn read_config(&mut self) -> Result<Box<Config>,String> {
        let f = open_file(self.path.to_string())?;
        from_reader(f)
            .map_err(|e| format!("Invalid Format: {}", e))
    }
}

fn open_file(path: String) -> Result<File,String> {
    File::open(&path).map_err(|e| format!("File open failed: {} on path {}", e, &path))	
}

#[cfg(test)]
mod tests {
    use super::*;

    macro_rules! test_path {
        ($arg:literal) => {
            concat!(env!("CARGO_MANIFEST_DIR"),"/src/resources/test/", $arg)
        };
    }

    fn test_config() -> Config {
        Config{
            name: "Axolotl-User".to_string(),
            tel: "+123".to_string(),
            uuid: Uuid::from_u128(0xa1a2a3a4b1b2c1c2d1d2d3d4d5d6d7d8u128),
            profile_key: vec![0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15],
            profile_key_credential: vec![0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15],
            certificate: vec![0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15]
        }
    }

    #[test]
    fn test_migrate_config_reading_fails() {
        struct TestReader();

        impl ConfigReader for TestReader {
            fn read_config(&mut self) -> Result<Box<Config>,String> {
                Err("test".to_string())
            }
        }
        let mut reader = TestReader();
        assert!(migrate(&mut reader).is_err());
    }

    #[test]
    fn test_migrate_config_success() {
        struct TestReader();

        impl ConfigReader for TestReader {
            fn read_config(&mut self) -> Result<Box<Config>,String> {
                let config = test_config();
                Ok(Box::new(config))
            }
        }
        let mut reader = TestReader();
        
        assert_eq!(migrate(&mut reader).unwrap(), ());
    }
    
    #[test]
    fn test_read_config_unknown_path_fails() {
        let mut reader = YamlConfigReader{ path: "some/path" };
        assert!(reader.read_config().is_err());
    }

    #[test]
    fn test_read_config_wrong_format_fails() {
        let mut reader = YamlConfigReader{ path: test_path!("test.json") };
        assert!(reader.read_config().is_err());
    }

    #[test]
    fn test_read_config_success() {
        let mut reader = YamlConfigReader{ path: test_path!("test_config.yml") };
        let config = test_config();
        assert_eq!(reader.read_config().unwrap(), Box::new(config));
    }
}