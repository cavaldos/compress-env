/// Encode the content to make it unreadable
pub fn encode_content(content: &str) -> Vec<u8> {
    let bytes = content.as_bytes();
    let mut encoded = Vec::with_capacity(bytes.len());
    
    // Simple XOR encoding with a fixed key
    let key: [u8; 8] = [0x42, 0x13, 0x37, 0x89, 0xAB, 0xCD, 0xEF, 0x01];
    
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
