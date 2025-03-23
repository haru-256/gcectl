use clap::Parser;
use log::debug;

#[derive(Debug, Parser)]
#[command(author, version, about, long_about=None)]
pub struct Args {
    // input text
    #[arg(value_name = "TEXT", help = "Input text", required = true)]
    text: Vec<String>,

    // omits the newline at the end of the output
    #[arg(
        short = 'n',
        long = "omit-newline",
        help = "Do not print newline",
        default_value_t = false
    )]
    omit_newline: bool,
}

fn main() {
    env_logger::init();

    let args = Args::parse();
    debug!("{:?}", args);

    let text = args.text;
    let omit_newline = args.omit_newline;

    let ending = if omit_newline { "" } else { "\n" };

    print!("{}", text.join(" ") + ending);
}
