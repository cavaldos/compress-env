use std::fs::{self, File};
use std::io::{self, Read, Write};
use std::path::Path;


pub fn compress_env_files() -> io::Result<()> {
    let current_dir = std::env::current_dir()?;
    let output_file_path = current_dir.join("env_contents.txt");
    let mut output_file = File::create(output_file_path)?;
    
    println!("Searching for .env files in: {}", current_dir.display());
    
    // Use a mutable String to collect all content
    let mut aggregated_content = String::new();
    
    // Find all .env files and collect their contents
    find_env_files(&current_dir, &mut aggregated_content)?;
    
    // Write the collected content to the output file
    output_file.write_all(aggregated_content.as_bytes())?;
    
    if aggregated_content.is_empty() {
        println!("No .env files found in the directory tree.");
    } else {
        println!("Aggregated .env contents saved to env_contents.txt");
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
                println!("Found .env file: {}", path.display());
                match File::open(&path) {
                    Ok(mut file) => {
                        let mut contents = String::new();
                        match file.read_to_string(&mut contents) {
                            Ok(_) => {
                                // Lấy đường dẫn tương đối từ thư mục hiện tại
                                let current_dir = std::env::current_dir()?;
                                let relative_path = path.strip_prefix(&current_dir).unwrap_or(&path);
                                
                                aggregated_content.push_str(&format!("--- {} ---\n", relative_path.display()));
                                aggregated_content.push_str(&contents);
                                aggregated_content.push_str("\n\n");
                            },
                            Err(e) => eprintln!("Error reading file {}: {}", path.display(), e),
                        }
                    },
                    Err(e) => eprintln!("Error opening file {}: {}", path.display(), e),
                }
            }
        }
    }
    Ok(())
}
