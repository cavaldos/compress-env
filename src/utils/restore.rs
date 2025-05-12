use std::fs::{self, File};
use std::io::{self, BufRead, BufReader, Write};
use std::path::{Path, PathBuf};


pub fn restore_env_files() -> io::Result<()> {
    let current_dir = std::env::current_dir()?;
    let input_file_path = current_dir.join("env_contents.txt");
    
    println!("Reading from: {}", input_file_path.display());
    
    let file = File::open(&input_file_path)?;
    let reader = BufReader::new(file);
    
    // Biến để theo dõi trạng thái phân tích
    let mut current_file: Option<PathBuf> = None;
    let mut current_content = String::new();
    let mut restored_files = Vec::new();
    
    // Đọc từng dòng của file
    for line in reader.lines() {
        let line = line?;
        
        // Kiểm tra nếu dòng bắt đầu bằng "--- "
        if line.starts_with("--- ") && line.ends_with(" ---") {
            // Lưu file hiện tại trước khi chuyển sang file mới
            if let Some(path) = current_file {
                if !current_content.is_empty() {
                    write_env_file(&path, &current_content)?;
                    restored_files.push(path);
                    current_content.clear();
                }
            }
            
            // Trích xuất tên file từ dòng
            let file_path = line.trim_start_matches("--- ").trim_end_matches(" ---");
            current_file = Some(current_dir.join(file_path));
        } else if current_file.is_some() {
            // Thêm dòng vào nội dung của file hiện tại
            if !line.is_empty() {
                current_content.push_str(&line);
                current_content.push('\n');
            }
        }
    }
    
    // Lưu file cuối cùng nếu có
    if let Some(path) = current_file {
        if !current_content.is_empty() {
            write_env_file(&path, &current_content)?;
            restored_files.push(path);
        }
    }
    
    // In thông báo kết quả
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

/// Hàm viết nội dung vào file .env
pub fn write_env_file(path: &Path, content: &str) -> io::Result<()> {
    // Tạo thư mục chứa file nếu chưa tồn tại
    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent)?;
    }
    
    // Viết nội dung vào file
    println!("Creating .env file: {}", path.display());
    let mut file = File::create(path)?;
    file.write_all(content.as_bytes())?;
    
    Ok(())
}
