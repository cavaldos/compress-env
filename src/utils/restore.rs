use std::fs::{self, File};
use std::io::{self, Read, Write};
use std::path::{Path, PathBuf};
use crate::utils::hash::decode::decode_content;


pub fn restore_env_files(input_file_name: Option<&str>, password: Option<&str>) -> io::Result<()> {
    let current_dir = std::env::current_dir()?;
    let filename = input_file_name.unwrap_or(crate::DEFAULT_ENV_FILE);
    let input_file_path = current_dir.join(filename);

    println!("Reading from: {}", input_file_path.display());

    // Read the entire file
    let mut file = File::open(&input_file_path)?;
    let mut encoded_data = Vec::new();
    file.read_to_end(&mut encoded_data)?;

    // Decode the content
    let decoded_content = match decode_content(&encoded_data, password) {
        Ok(content) => content,
        Err(e) => {
            eprintln!("Error decoding file: {}", e);
            return Err(e);
        }
    };

    // Process the decoded content
    let mut current_file: Option<PathBuf> = None;
    let mut current_content = String::new();
    let mut restored_files = Vec::new();

    // Process each line of the decoded content
    for line in decoded_content.lines() {
        // Check if line starts with "--- " and ends with " ---"
        if line.starts_with("--- ") && line.ends_with(" ---") {
            // Save current file before moving to the next one
            if let Some(path) = current_file {
                if !current_content.is_empty() {
                    write_env_file(&path, &current_content)?;
                    restored_files.push(path);
                    current_content.clear();
                }
            }

            // Extract file path from the line
            let file_path = line.trim_start_matches("--- ").trim_end_matches(" ---");
            current_file = Some(current_dir.join(file_path));
        } else if current_file.is_some() {
            // Add line to current file content
            if !line.is_empty() {
                current_content.push_str(line);
                current_content.push('\n');
            }
        }
    }

    // Save the last file if any
    if let Some(path) = current_file {
        if !current_content.is_empty() {
            write_env_file(&path, &current_content)?;
            restored_files.push(path);
        }
    }

    // Print result message
    if restored_files.is_empty() {
        println!("No .env files were restored.");
    } else {
        println!("Restored {} .env files:", restored_files.len());
        for file in restored_files {
            println!("  - {}", file.display());
        }
    }

    Ok(())
}

/// Write content to an .env file
pub fn write_env_file(path: &Path, content: &str) -> io::Result<()> {
    // Create directory if it doesn't exist
    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent)?;
    }

    // Write content to file
    println!("Creating .env file: {}", path.display());
    let mut file = File::create(path)?;
    file.write_all(content.as_bytes())?;

    Ok(())
}


