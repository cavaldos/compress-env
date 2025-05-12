pub fn print_usage() {
    println!("Usage: compress-env [COMMAND] [OPTIONS]");
    println!("Commands:");
    println!("  compress          - Compress all .env files into a single backup file");
    println!("  restore           - Restore .env files from the backup file");
    println!("Options:"); 
    println!("  -o <filename>     - Specify the output filename for compression (default: {})", crate::DEFAULT_ENV_FILE);
    println!("  -i <filename>     - Specify the input filename for restoration (default: {})", crate::DEFAULT_ENV_FILE);
    println!("  -f <file1> <file2> ... - Specify specific .env files to compress");
    println!("  -h, --help        - Show this help message");
    println!("  -v, --version     - Show version information");
    println!("Examples:");
    println!("  compress-env compress -o backup.env");
    println!("  compress-env compress -f ./project1/.env ./project2/.env");
    println!("  compress-env compress -f ./project1/.env -o custom-backup.env");
    println!("  compress-env restore -i backup.env");
}