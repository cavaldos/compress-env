pub fn print_usage() {
    println!("Usage: compress-env [COMMAND] [OPTIONS]");
    println!("Commands:");
    println!("  compress          - Compress and encode all .env files into a single encrypted binary file");
    println!("  restore           - Decode and restore .env files from the encrypted binary file");
    println!("Options:");
    println!("  -o <filename>     - Specify the output filename for compression (default: {})", crate::DEFAULT_ENV_FILE);
    println!("  -i <filename>     - Specify the input filename for restoration (default: {})", crate::DEFAULT_ENV_FILE);
    println!("  -f <file1> <file2> ... - Specify specific .env files to compress");
    println!("  -p <password>     - Specify a password for encryption/decryption (optional)");
    println!("  -h, --help        - Show this help message");
    println!("  -v, --version     - Show version information");
    println!("Examples:");
    println!("  compress-env compress -o secrets.bin");
    println!("  compress-env compress -f ./project1/.env -o project1-secrets.bin -p mypassword");
    println!("  compress-env restore -i secrets.bin -p mypassword");
}