/* Copyright (C) 2018 Olivier Goffart <ogoffart@woboq.com>
Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:
The above copyright notice and this permission notice shall be included in all copies or substantial
portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES
OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/
use std::env;
use std::io::prelude::*;
use std::io::BufReader;
use std::path::Path;
use std::process::Command;

use failure::*;
use vergen::*;

fn qmake_query(var: &str) -> String {
    let qmake = std::env::var("QMAKE").unwrap_or_else(|_| "qmake".to_string());
    String::from_utf8(
        Command::new(qmake)
            .env("QT_SELECT", "qt5")
            .args(&["-query", var])
            .output()
            .expect("Failed to execute qmake. Make sure 'qmake' is in your path")
            .stdout,
    )
    .expect("UTF-8 conversion failed")
}

fn mock_pthread(mer_root: &str, arch: &str) -> Result<String, Error> {
    let out_dir = env::var("OUT_DIR")?;
    let qml_path = &Path::new(&out_dir).join("libpthread.so");

    let mut f = std::fs::File::create(qml_path)?;
    match arch {
        "armv7hl" => {
            writeln!(f, "OUTPUT_FORMAT(elf32-littlearm)")?;
        }
        "i486" => {
            writeln!(f, "OUTPUT_FORMAT(elf32-i386)")?;
        }
        "aarch64" => {
            writeln!(f, "OUTPUT_FORMAT(elf64-littleaarch64)")?;
        }
        _ => unreachable!(),
    }

    match arch {
        "armv7hl" | "i486" => writeln!(
            f,
            "GROUP ( {}/lib/libpthread.so.0 {}/usr/lib/libpthread_nonshared.a )",
            mer_root, mer_root
        )?,
        "aarch64" => writeln!(
            f,
            "GROUP ( {}/lib64/libpthread.so.0 {}/usr/lib64/libpthread_nonshared.a )",
            mer_root, mer_root
        )?,
        _ => unreachable!(),
    }

    Ok(out_dir)
}

fn mock_libc(mer_root: &str, arch: &str) -> Result<String, Error> {
    let out_dir = env::var("OUT_DIR")?;
    let qml_path = &Path::new(&out_dir).join("libc.so");

    let mut f = std::fs::File::create(qml_path)?;
    match arch {
        "armv7hl" => {
            writeln!(f, "OUTPUT_FORMAT(elf32-littlearm)")?;
            writeln!(f, "GROUP ( {}/lib/libc.so.6 {}/usr/lib/libc_nonshared.a  AS_NEEDED ( {}/lib/ld-linux-armhf.so.3 ))",
                mer_root, mer_root, mer_root)?;
        }
        "i486" => {
            writeln!(f, "OUTPUT_FORMAT(elf32-i386)")?;
            writeln!(f, "GROUP ( {}/lib/libc.so.6 {}/usr/lib/libc_nonshared.a  AS_NEEDED ( {}/lib/ld-linux.so.2 ))",
                mer_root, mer_root, mer_root)?;
        }
        "aarch64" => {
            writeln!(f, "OUTPUT_FORMAT(elf64-littleaarch64)")?;
            writeln!(f, "GROUP ( {}/lib64/libc.so.6 {}/usr/lib64/libc_nonshared.a  AS_NEEDED ( {}/lib64/ld-linux-aarch64.so.1 ))",
                mer_root, mer_root, mer_root)?;
        }
        _ => unreachable!(),
    }

    Ok(out_dir)
}

fn install_mer_hacks() -> (String, bool) {
    let mer_sdk = match std::env::var("MERSDK").ok() {
        Some(path) => path,
        None => return ("".into(), false),
    };

    let mer_target = std::env::var("MER_TARGET")
        .ok()
        .unwrap_or_else(|| "SailfishOS-latest".into());

    let arch = match &std::env::var("CARGO_CFG_TARGET_ARCH").unwrap() as &str {
        "arm" => "armv7hl",
        "i686" => "i486",
        "x86" => "i486",
        "aarch64" => "aarch64",
        unsupported => panic!("Target {} is not supported for Mer", unsupported),
    };

    let lib_dir = match arch {
        "armv7hl" | "i486" => "lib",
        "aarch64" => "lib64",
        _ => unreachable!(),
    };

    println!("cargo:rustc-cfg=feature=\"sailfish\"");

    let mer_target_root = format!("{}/targets/{}-{}", mer_sdk, mer_target, arch);

    let mock_libc_path = mock_libc(&mer_target_root, arch).unwrap();
    let mock_pthread_path = mock_pthread(&mer_target_root, arch).unwrap();

    let macos_lib_search = if cfg!(target_os = "macos") {
        "=framework"
    } else {
        ""
    };

    println!(
        "cargo:rustc-link-search{}={}",
        macos_lib_search, mock_pthread_path,
    );
    println!(
        "cargo:rustc-link-search{}={}",
        macos_lib_search, mock_libc_path,
    );

    println!(
        "cargo:rustc-link-arg-bins=-rpath-link,{}/usr/{}",
        mer_target_root, lib_dir
    );
    println!(
        "cargo:rustc-link-arg-bins=-rpath-link,{}/{}",
        mer_target_root, lib_dir
    );

    println!(
        "cargo:rustc-link-search{}={}/toolings/{}/opt/cross/{}-meego-linux-gnueabi/{}",
        macos_lib_search, mer_sdk, mer_target, arch, lib_dir
    );

    println!(
        "cargo:rustc-link-search{}={}/usr/{}/qt5/qml/Nemo/Notifications/",
        macos_lib_search, mer_target_root, lib_dir
    );

    println!(
        "cargo:rustc-link-search{}={}/toolings/{}/opt/cross/{}/gcc/{}-meego-linux-gnueabi/4.9.4/",
        macos_lib_search, mer_sdk, mer_target, arch, lib_dir
    );

    println!(
        "cargo:rustc-link-search{}={}/usr/{}/",
        macos_lib_search, mer_target_root, lib_dir
    );

    (mer_target_root, true)
}


fn protobuf() -> Result<(), Error> {
    let protobuf = Path::new("protobuf").to_owned();

    let input: Vec<_> = protobuf
        .read_dir()
        .expect("protobuf directory")
        .filter_map(|entry| {
            let entry = entry.expect("readable protobuf directory");
            let path = entry.path();
            if Some("proto") == path.extension().and_then(std::ffi::OsStr::to_str) {
                assert!(path.is_file());
                println!("cargo:rerun-if-changed={}", path.to_str().unwrap());
                Some(path)
            } else {
                None
            }
        })
        .collect();

    prost_build::compile_protos(&input, &[protobuf])?;
    Ok(())
}

fn prepare_rpm_build() {
    println!("cargo:rerun-if-env-changed=CARGO_FEATURE_HARBOUR");

    // Adding files only for a specific feature:
    // - Register a new folder in cond_folder and add .rpm/tmp_feature_files/new_folder to
    //   Cargo.toml be included in the rpm where your file should end up.
    // - Add your source file to cond_files as ("filename", "new_folder", generated) with gerenated
    //   = true if the file is generated during build (this prevents a rerun-if-changeged line to be
    //   emitted)
    // - Add your file to the rpm spec between lines `#[{{ [NOT] FEATURE_FLAG` and `#}}]`

    let rpm_extra_dir = std::path::PathBuf::from(".rpm/tmp_feature_files");
    if rpm_extra_dir.exists() {
        std::fs::remove_dir_all(&rpm_extra_dir)
            .unwrap_or_else(|_| panic!("Could not remove {:?} for cleanup", rpm_extra_dir));
    }
    let cond_folder: &[&str] = &["systemd", "transferplugin", "transferui", "dbus"];
    for d in cond_folder.iter() {
        let nd = rpm_extra_dir.join(d);
        std::fs::create_dir_all(&nd).unwrap_or_else(|_| panic!("Could not create {:?}", &nd));
    }
    let cond_files: &[(&str, &str, bool)] = if env::var("CARGO_FEATURE_HARBOUR").is_err() {
        // [ ("file name", "destination folder", generated), ... ]
        &[
            ("harbour-whisperfish.service", "systemd", false),
            ("be.rubdos.whisperfish.service", "dbus", false),
            (
                "shareplugin/libwhisperfishshareplugin.so",
                "transferplugin",
                true,
            ),
            ("shareplugin/WhisperfishShare.qml", "transferui", false),
        ]
    } else {
        &[]
    };
    for (file, dest, gen) in cond_files.iter() {
        let dest_dir = rpm_extra_dir.join(dest);
        if !dest_dir.exists() {
            std::fs::create_dir_all(&dest_dir)
                .unwrap_or_else(|_| panic!("Could not create {:?}", dest_dir));
        }
        let dest_file = dest_dir.join(Path::new(file).file_name().unwrap());
        std::fs::copy(Path::new(file), &dest_file)
            .unwrap_or_else(|_| panic!("failed to copy {} to {:?}", file, dest_file));
        if !gen {
            println!("cargo:rerun-if-changed={}", file);
        }
    }

    // Build RPM Spec
    // Lines between `#[{{ NOT FEATURE_FLAG` and `#}}]` are only copied if the feature is disabled
    // (or enabled without NOT).
    println!("cargo:rerun-if-changed=rpm/harbour-whisperfish.spec");
    let src = std::fs::File::open("rpm/harbour-whisperfish.spec")
        .expect("Failed to read rpm spec at rpm/harbour-whisperfish.spec");
    let mut spec = std::fs::File::create(".rpm/harbour-whisperfish.spec")
        .expect("Failed to write rpm spec to .rpm/harbour-whisperfish.spec");
    writeln!(spec, "### WARNING: auto-generated file - please only edit the original source file: ../rpm/harbour-whisperfish.spec")
        .expect("Failed to write to spec file");

    let mut ignore = 0;
    let feature_re = regex::Regex::new(r"^\s*#\[\{\{\s+(NOT)?\s+([A-Z_0-9]+)").unwrap();

    for line in BufReader::new(src).lines() {
        let line = line.unwrap();
        if let Some(cap) = feature_re.captures(&line) {
            if ignore > 0
                || (cap.get(1) == None
                    && env::var(format!("CARGO_FEATURE_{}", cap.get(2).unwrap().as_str())).is_err())
                || (cap.get(1) != None
                    && env::var(format!("CARGO_FEATURE_{}", cap.get(2).unwrap().as_str())).is_ok())
            {
                ignore += 1;
            }
            println!("reg {:?}", cap);
        } else if line.trim_start().starts_with("#}}]") {
            if ignore > 0 {
                ignore -= 1;
            }
        } else if ignore == 0 {
            writeln!(spec, "{}", line).expect("Failed to write to spec file");
        }
    }
}

fn needs_rerun(dest: &str, sources: &[&str]) -> bool {
    for f in sources.iter() {
        println!("cargo:rerun-if-changed={}", f);
    }

    let metadata = std::fs::metadata(dest);
    if metadata.is_err() {
        return true;
    }
    let reftime = match metadata.unwrap().modified() {
        Ok(time) => time,
        Err(_) => return true,
    };

    for f in sources.iter() {
        if std::fs::metadata(f).unwrap().modified().unwrap() > reftime {
            return true;
        }
    }

    false
}

fn build_share_plugin(mer_target_root: &str, qt_include_path: &str) {
    if !needs_rerun(
        "shareplugin/libwhisperfishshareplugin.so",
        &[
            "shareplugin/WhisperfishPluginInfo.cpp",
            "shareplugin/WhisperfishPluginInfo.h",
            "shareplugin/WhisperfishTransferPlugin.cpp",
            "shareplugin/sfmoc/WhisperfishTransferPlugin.cpp",
            "shareplugin/WhisperfishTransferPlugin.h",
            "shareplugin/WhisperfishTransfer.cpp",
            "shareplugin/sfmoc/WhisperfishTransfer.cpp",
            "shareplugin/WhisperfishTransfer.h",
        ],
    ) {
        return;
    }

    let mut gcc = cc::Build::new()
        .cargo_metadata(false)
        .cpp(true)
        .flag(&format!("--sysroot={}", mer_target_root))
        .flag("-isysroot")
        .flag(mer_target_root)
        .include(format!("{}/usr/include/", mer_target_root))
        .include(&qt_include_path)
        .include(format!("{}/QtCore", qt_include_path))
        .shared_flag(true)
        .get_compiler()
        .to_command();
    gcc.arg(format!("-L{}/usr/lib64", mer_target_root))
        .arg(format!("-L{}/usr/lib", mer_target_root))
        .arg("-lnemotransferengine-qt5")
        .arg("-lQt5DBus")
        .arg("-lQt5Core")
        .arg("-o")
        .arg("shareplugin/libwhisperfishshareplugin.so")
        .arg("shareplugin/WhisperfishPluginInfo.cpp")
        .arg("shareplugin/WhisperfishTransfer.cpp")
        .arg("shareplugin/WhisperfishTransferPlugin.cpp")
        .arg("shareplugin/sfmoc/WhisperfishTransfer.cpp")
        .arg("shareplugin/sfmoc/WhisperfishTransferPlugin.cpp");

    println!("running: {:?}", gcc);
    gcc.status().expect("share plugin compile command failed");
}

fn build_sqlcipher(mer_target_root: &str) {
    // static sqlcipher handling. Needed for compatibility with
    // sailfish-components-webview.
    // This may become obsolete with an sqlcipher upgrade from jolla or when
    // https://gitlab.com/rubdos/whisperfish/-/issues/227 is implemented.

    if !needs_rerun(
        &format!("{}/libsqlcipher.a", env::var("OUT_DIR").unwrap()),
        &[
            "sqlcipher/sqlite3.c",
            "sqlcipher/sqlite3.h",
            "sqlcipher/sqlite3ext.h",
        ],
    ) {
        return;
    }

    if !Path::new("sqlcipher/sqlite3.c").is_file() {
        // Download and prepare sqlcipher source
        let stat = Command::new("sqlcipher/get-sqlcipher.sh")
            .status()
            .expect("Failed to download sqlcipher");
        assert!(stat.success());
    }

    // Build static sqlcipher
    cc::Build::new()
        .flag(&format!("--sysroot={}", mer_target_root))
        .flag("-isysroot")
        .flag(mer_target_root)
        .include(format!("{}/usr/include/", mer_target_root))
        .include(format!("{}/usr/include/openssl", mer_target_root))
        .file("sqlcipher/sqlite3.c")
        .warnings(false)
        .flag("-Wno-stringop-overflow")
        .flag("-Wno-return-local-addr")
        .flag("-DSQLITE_CORE")
        .flag("-DSQLITE_DEFAULT_FOREIGN_KEYS=1")
        .flag("-DSQLITE_ENABLE_API_ARMOR")
        .flag("-DSQLITE_HAS_CODEC")
        .flag("-DSQLITE_TEMP_STORE=2")
        .flag("-DHAVE_ISNAN")
        .flag("-DHAVE_LOCALTIME_R")
        .flag("-DSQLITE_ENABLE_COLUMN_METADATA")
        .flag("-DSQLITE_ENABLE_DBSTAT_VTAB")
        .flag("-DSQLITE_ENABLE_FTS3")
        .flag("-DSQLITE_ENABLE_FTS3_PARENTHESIS")
        .flag("-DSQLITE_ENABLE_FTS5")
        .flag("-DSQLITE_ENABLE_JSON1")
        .flag("-DSQLITE_ENABLE_LOAD_EXTENSION=1")
        .flag("-DSQLITE_ENABLE_MEMORY_MANAGEMENT")
        .flag("-DSQLITE_ENABLE_RTREE")
        .flag("-DSQLITE_ENABLE_STAT2")
        .flag("-DSQLITE_ENABLE_STAT4")
        .flag("-DSQLITE_SOUNDEX")
        .flag("-DSQLITE_THREADSAFE=1")
        .flag("-DSQLITE_USE_URI")
        .flag("-DHAVE_USLEEP=1")
        .compile("sqlcipher");

    println!("cargo:lib_dir={}", env::var("OUT_DIR").unwrap());
    println!("cargo:rustc-link-lib=static=sqlcipher");
}

fn main() {
    protobuf().unwrap();

    // Print a warning when rustc is too old.
    if !version_check::is_min_version("1.48.0").unwrap_or(false) {
        if let Some(version) = version_check::Version::read() {
            panic!(
                "Whisperfish requires Rust 1.48.0 or later.  You are using rustc {}",
                version
            );
        } else {
            panic!(
                "Whisperfish requires Rust 1.48.0 or later, but could not determine Rust version.",
            );
        }
    }

    let (mer_target_root, cross_compile) = install_mer_hacks();
    let qt_include_path = if cross_compile {
        format!("{}/usr/include/qt5/", mer_target_root)
    } else {
        qmake_query("QT_INSTALL_HEADERS")
    };
    let qt_include_path = qt_include_path.trim();

    let mut cfg = cpp_build::Config::new();

    cfg.flag(&format!("--sysroot={}", mer_target_root));
    cfg.flag("-isysroot");
    cfg.flag(&mer_target_root);

    // https://github.com/rust-lang/cargo/pull/8441/files
    // currently requires -Zextra-link-arg, so we're duplicating this in dotenv
    println!("cargo:rustc-link-arg=--sysroot={}", mer_target_root);


    // This is kinda hacky. Sorry.
    if cross_compile {
        std::env::set_var("CARGO_FEATURE_SAILFISH", "");
    }
    cfg.include(format!("{}/usr/include/sailfishapp/", mer_target_root))
        .include(&qt_include_path)
        .include(format!("{}/QtCore", qt_include_path))
        // It is annoying to look at while developing, and we cannot do anything about it
        // ourselves.
        .flag("-Wno-deprecated-copy")
        .build("src/main.rs");

    let contains_cpp = [
        "qmlapp/mod.rs",
        "qmlapp/tokio_qt.rs",
        "qmlapp/native.rs",
        "config/settings.rs",
    ];
    for f in &contains_cpp {
        println!("cargo:rerun-if-changed=src/{}", f);
    }

    let macos_lib_search = if cfg!(target_os = "macos") {
        "=framework"
    } else {
        ""
    };

    let sailfish_libs: &[&str] = if cross_compile {
        &["nemonotifications", "sailfishapp", "qt5embedwidget"]
    } else {
        &[]
    };
    let libs = ["EGL", "dbus-1"];
    for lib in libs.iter().chain(sailfish_libs.iter()) {
        println!("cargo:rustc-link-lib{}={}", macos_lib_search, lib);
    }

    if env::var("CARGO_FEATURE_HARBOUR").is_err() && cross_compile {
        build_share_plugin(&mer_target_root, qt_include_path);
    }

    if cross_compile {
        build_sqlcipher(&mer_target_root);
    }

    if cross_compile {
        prepare_rpm_build();
    }

    // vergen
    let mut cfg = vergen::Config::default();
    *cfg.git_mut().enabled_mut() = true;
    *cfg.git_mut().sha_mut() = true;
    *cfg.git_mut().sha_kind_mut() = vergen::ShaKind::Short;
    vergen(cfg).expect("vergen setup");
}
