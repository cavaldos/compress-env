use std::fs::{self, File};
use std::io::{self, Read, Write};
use std::path::{Path, PathBuf};
use crate::utils::hash::encode::encode_content;


pub fn compress_env_files(output_file_name: Option<&str>, specific_files: Option<Vec<&str>>, password: Option<&str>) -> io::Result<()> {
    let current_dir = std::env::current_dir()?;
    let filename = output_file_name.unwrap_or(crate::DEFAULT_ENV_FILE);
    let output_file_path = current_dir.join(filename);
    let mut output_file = File::create(&output_file_path)?;

    // Use a mutable String to collect all content
    let mut aggregated_content = String::new();

    match specific_files {
        Some(files) => {
            println!("Compressing specified .env files...");
            for file_path in files {
                let path = PathBuf::from(file_path);
                if path.exists() && path.is_file() {
                    process_env_file(&path, &mut aggregated_content)?;
                } else {
                    eprintln!("Warning: File not found or not a regular file: {}", file_path);
                }
            }
        },
        None => {
            println!("Searching for .env files in: {}", current_dir.display());
            // Find all .env files and collect their contents
            find_env_files(&current_dir, &mut aggregated_content)?;
        }
    };

    // Encode the content before writing to file
    let encoded_content = encode_content(&aggregated_content, password);

    // Write the encoded content to the output file
    output_file.write_all(&encoded_content)?;

    if aggregated_content.is_empty() {
        println!("No .env files found in the directory tree.");
    } else {
        println!("Encoded .env contents saved to {}", output_file_path.display());
    }

    Ok(())
}

/// Find all .env files and collect their contents
pub fn find_env_files(dir: &Path, aggregated_content: &mut String) -> io::Result<()> {
    if dir.is_dir() {
        for entry in fs::read_dir(dir)? {
            let entry = entry?;
            let path = entry.path();

            if path.is_dir() {
                find_env_files(&path, aggregated_content)?;
            } else if path.is_file() && path.file_name().map_or(false, |name| name == ".env") {
                process_env_file(&path, aggregated_content)?;
            }
        }
    }
    Ok(())
}

/// Process a single env file and add its contents to the aggregated content
pub fn process_env_file(path: &Path, aggregated_content: &mut String) -> io::Result<()> {
    match File::open(path) {
        Ok(mut file) => {
            let mut contents = String::new();
            match file.read_to_string(&mut contents) {
                Ok(_) => {
                    // Get the relative path from the current directory
                    let current_dir = std::env::current_dir()?;
                    let relative_path = path.strip_prefix(&current_dir).unwrap_or(path);

                    aggregated_content.push_str(&format!("--- {} ---\n", relative_path.display()));
                    aggregated_content.push_str(&contents);
                    aggregated_content.push_str("\n\n");
                    Ok(())
                },
                Err(e) => {
                    eprintln!("Error reading file {}: {}", path.display(), e);
                    Err(e)
                }
            }
        },
        Err(e) => {
            eprintln!("Error opening file {}: {}", path.display(), e);
            Err(e)
        }
    }
}


