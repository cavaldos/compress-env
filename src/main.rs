mod utils;
use utils::compress::compress_env_files;
use utils::restore::restore_env_files;

fn main() -> std::io::Result<()> {
    println!("Compressing .env files");
    restore_env_files()?;
    Ok(())
}
