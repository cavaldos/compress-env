use std::io;
use std::collections::hash_map::DefaultHasher;
use std::hash::{Hash, Hasher};

/// Decode the binary content back to readable text
pub fn decode_content(encoded_data: &[u8], password: Option<&str>) -> io::Result<String> {
    // Check if the file has the correct signature
    if encoded_data.len() < 12 || &encoded_data[0..7] != b"ENVENC0" {
        return Err(io::Error::new(
            io::ErrorKind::InvalidData,
            "Invalid file format: not an encoded .env file"
        ));
    }

    // Check version
    if encoded_data[7] != b'1' {
        return Err(io::Error::new(
            io::ErrorKind::InvalidData,
            format!("Unsupported version: {}", encoded_data[7] as char)
        ));
    }

    // Read the length
    let mut length_bytes = [0u8; 4];
    length_bytes.copy_from_slice(&encoded_data[8..12]);
    let length = u32::from_le_bytes(length_bytes) as usize;

    // Ensure the file contains enough data
    if encoded_data.len() < 12 + length {
        return Err(io::Error::new(
            io::ErrorKind::InvalidData,
            "File is truncated or corrupted"
        ));
    }

    // Extract the encoded content
    let encoded_content = &encoded_data[12..12 + length];

    // Generate key from password or use default key
    let key = generate_key_from_password(password);
    let mut decoded = Vec::with_capacity(encoded_content.len());

    for (i, &byte) in encoded_content.iter().enumerate() {
        // XOR with the same key byte
        let decoded_byte = byte ^ key[i % key.len()];
        decoded.push(decoded_byte);
    }

    // Convert back to string
    match String::from_utf8(decoded) {
        Ok(s) => Ok(s),
        Err(e) => Err(io::Error::new(
            io::ErrorKind::InvalidData,
            format!("Invalid UTF-8 sequence or incorrect password: {}", e)
        )),
    }
}

/// Generate a key from the provided password or use default key if none
fn generate_key_from_password(password: Option<&str>) -> Vec<u8> {
    match password {
        Some(pass) if !pass.is_empty() => {
            // Use password to generate a key
            let mut key = Vec::with_capacity(16);
            let mut hasher = DefaultHasher::new();
            pass.hash(&mut hasher);
            let hash = hasher.finish();

            // Use the hash to generate a 16-byte key
            key.extend_from_slice(&hash.to_le_bytes());

            // Add more bytes based on variations of the password
            let mut hasher = DefaultHasher::new();
            (pass.to_string() + "salt1").hash(&mut hasher);
            key.extend_from_slice(&hasher.finish().to_le_bytes());

            key
        },
        _ => {
            // Default key if no password provided
            vec![0x42, 0x13, 0x37, 0x89, 0xAB, 0xCD, 0xEF, 0x01,
                 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF, 0x01]
        }
    }
}
