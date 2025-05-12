use std::io;

/// Decode the binary content back to readable text
pub fn decode_content(encoded_data: &[u8]) -> io::Result<String> {
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
    
    // Decode using the same key as in encode_content
    let key: [u8; 8] = [0x42, 0x13, 0x37, 0x89, 0xAB, 0xCD, 0xEF, 0x01];
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
            format!("Invalid UTF-8 sequence: {}", e)
        )),
    }
}
