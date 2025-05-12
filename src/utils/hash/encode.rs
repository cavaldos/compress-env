use std::collections::hash_map::DefaultHasher;
use std::hash::{Hash, Hasher};

/// Encode the content to make it unreadable
pub fn encode_content(content: &str, password: Option<&str>) -> Vec<u8> {
    let bytes = content.as_bytes();
    let mut encoded = Vec::with_capacity(bytes.len());

    // Generate key from password or use default key
    let key = generate_key_from_password(password);

    for (i, &byte) in bytes.iter().enumerate() {
        // XOR with key byte (cycling through the key)
        let encoded_byte = byte ^ key[i % key.len()];
        encoded.push(encoded_byte);
    }

    // Add a signature at the beginning to identify this as an encoded file
    let mut result = Vec::new();
    result.extend_from_slice(b"ENVENC01"); // Signature/version
    result.extend_from_slice(&(encoded.len() as u32).to_le_bytes()); // Length
    result.extend(encoded);

    result
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
