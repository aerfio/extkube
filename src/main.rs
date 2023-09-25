use anyhow::{Context, Result};
use arboard::Clipboard;
use clap::Parser;
use colored::Colorize;
use kube::config::Kubeconfig;
use tempfile::Builder;

/// Application to extract selected kubeconfig context from active kubeconfig file.
#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// k8s context
    #[arg(short, long)]
    context: String,
}

fn main() -> Result<()> {
    let mut clipboard = Clipboard::new()?;
    let ctx = Args::parse().context;
    let mut kubecfg = Kubeconfig::read()?;
    kubecfg.current_context = Some(ctx.clone());

    let selected_context = kubecfg
        .contexts
        .iter()
        .find(|named_context| ctx == named_context.name)
        .with_context(|| format!("context {} not found in active kubeconfig", ctx))?;

    // let cluster_name = cluster.unwrap_or(&current_context.cluster);
    let cluster = kubecfg
        .clusters
        .iter()
        .find(|&named_cluster| {
            selected_context
                .context
                .clone()
                .is_some_and(|val| val.cluster == named_cluster.name)
        })
        .with_context(|| format!("cluster not found for context {} in active kubeconfig", ctx))?;

    let user = kubecfg
        .auth_infos
        .iter()
        .find(|named_user| {
            selected_context
                .clone()
                .context
                .is_some_and(|val| val.user == named_user.name)
        })
        .with_context(|| format!("user not found for context {} in active kubeconfig", ctx))?;

    kubecfg.contexts = vec![selected_context.clone()];
    kubecfg.auth_infos = vec![user.clone()];
    kubecfg.clusters = vec![cluster.clone()];

    let named_tempfile = Builder::new().prefix("kubeconfig-").tempfile()?.keep()?;
    serde_yaml::to_writer(named_tempfile.0, &kubecfg)?;
    clipboard.set_text(format!(
        "export KUBECONFIG=\"{}\"",
        named_tempfile.1.to_string_lossy()
    ))?;
    println!("Copied:");
    println!(
        "{}",
        format!(
            "export KUBECONFIG=\"{}\"",
            named_tempfile
                .1
                .to_str()
                .with_context(|| "could not convert path of temporary file to string")?
        )
        .bright_green()
    );
    println!("to clipboard");
    Ok(())
}
