mod utils;
use std::env;
use utils::compress::compress_env_files;
use utils::restore::restore_env_files;
use utils::help::print_usage;

// Define the default filename as a constant
pub const DEFAULT_ENV_FILE: &str = ".env.bin";

fn main() -> std::io::Result<()> {
    let args: Vec<String> = env::args().collect();
    
    if args.len() <= 1 {
        print_usage();
        return Ok(());
    }

    match args[1].as_str() {
        "compress" => {
            println!("Compressing .env files...");
            
            // Parse command line options
            let mut output_file = None;
            let mut specific_files = Vec::new();
            let mut i = 2;
            
            while i < args.len() {
                match args[i].as_str() {
                    "-o" => {
                        if i + 1 < args.len() {
                            output_file = Some(args[i + 1].as_str());
                            i += 2;
                        } else {
                            eprintln!("Error: -o option requires a filename");
                            print_usage();
                            return Ok(());
                        }
                    },
                    "-f" => {
                        i += 1;
                        while i < args.len() && !args[i].starts_with('-') {
                            specific_files.push(args[i].as_str());
                            i += 1;
                        }
                    },
                    _ => {
                        if args[i].starts_with('-') {
                            eprintln!("Unknown option: {}", args[i]);
                            print_usage();
                            return Ok(());
                        }
                        i += 1;
                    }
                }
            }
            
            let specific_files_option = if specific_files.is_empty() {
                None
            } else {
                Some(specific_files)
            };
            
            compress_env_files(output_file, specific_files_option)?;
        },
        "restore" => {
            println!("Restoring .env files from backup...");
            
            // Check for -i option
            let mut input_file = None;
            let mut i = 2;
            while i < args.len() {
                if args[i] == "-i" && i + 1 < args.len() {
                    input_file = Some(args[i + 1].as_str());
                    i += 2;
                } else {
                    i += 1;
                }
            }
            
            restore_env_files(input_file)?;
        },
        _ => {
            print_usage();
        }
    }
    
    Ok(())
}
