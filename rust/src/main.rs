use anyhow::Result;
use log::info;
use regex::Regex;
use reqwest::header;
use serde::ser::SerializeStruct;
use std::default::Default;
use std::path::Path;
use std::process::Command;
use std::time::{Duration, Instant};
use std::{env, fs};

struct OpenAIClient {
    client: reqwest::blocking::Client,
    base_url: reqwest::Url,
}

#[derive(serde::Deserialize)]
struct Config {
    #[serde(default = "default_trace")]
    trace: bool,
    #[serde(default)]
    provider: ProviderConfig,
    #[serde(default)]
    diff: DiffConfig,
}

fn default_trace() -> bool {
    false
}

#[derive(serde::Deserialize)]
struct ProviderConfig {
    #[serde(default = "default_base_url")]
    base_url: String,
    #[serde(default = "default_api_key")]
    api_key: String,
}

impl Default for ProviderConfig {
    fn default() -> Self {
        Self {
            base_url: default_base_url(),
            api_key: default_api_key(),
        }
    }
}

fn default_base_url() -> String {
    "https://generativelanguage.googleapis.com/v1beta/openai/".to_string()
}

fn default_api_key() -> String {
    "GEMINI_API_KEY".to_string()
}

#[derive(serde::Deserialize)]
struct DiffConfig {
    #[serde(default = "default_short_model")]
    short_model: String,
    #[serde(default = "default_long_model")]
    long_model: String,
    #[serde(default = "default_threshold")]
    threshold: u32,
}

impl Default for DiffConfig {
    fn default() -> Self {
        Self {
            short_model: default_short_model(),
            long_model: default_long_model(),
            threshold: default_threshold(),
        }
    }
}

fn default_short_model() -> String {
    "gemini-2.5-flash-lite".to_string()
}

fn default_long_model() -> String {
    "gemini-2.5-flash".to_string()
}

fn default_threshold() -> u32 {
    200
}

#[derive(serde::Serialize)]
struct Chat {
    model: String,
    messages: Vec<ChatMessage>,
}

#[derive(serde::Serialize)]
struct ChatMessage {
    role: String,
    content: String,
}

#[derive(serde::Deserialize)]
struct Completion {
    choices: Vec<Choice>,
}

#[derive(serde::Deserialize)]
struct Choice {
    message: ChoiceMessage,
}

#[derive(serde::Deserialize)]
struct ChoiceMessage {
    role: String,
    content: String,
}

struct TraceWrapper(Trace);

impl serde::Serialize for TraceWrapper {
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let mut s = serializer.serialize_struct("TraceWrapper", 1)?;
        s.serialize_field("auto-commit-msg", &self.0)?;
        s.end()
    }
}

#[derive(serde::Serialize)]
struct Trace {
    language: String,
    model: String,
    response_time: TraceDuration,
    execution_time: TraceDuration,
}

struct TraceDuration(Duration);

impl serde::Serialize for TraceDuration {
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let secs = self.0.as_secs_f64();
        serializer.serialize_f64(secs)
    }
}

impl OpenAIClient {
    fn build(base_url: reqwest::Url, token: String) -> Result<Self> {
        let mut headers = header::HeaderMap::new();
        headers.insert(header::AUTHORIZATION, format!("Bearer {token}").parse()?);
        let client = reqwest::blocking::Client::builder()
            .default_headers(headers)
            .build()?;
        Ok(Self { client, base_url })
    }

    fn create_chat_completion(&self, chat: Chat) -> Result<Completion> {
        let url = self.base_url.join("chat/completions")?;
        let completion = self.client.post(url).json(&chat).send()?.json()?;
        Ok(completion)
    }
}

fn main() -> Result<()> {
    let execution_duration = Instant::now();

    let mut config_content = "".to_string();
    if Path::new(".auto-commit-msg.toml").exists() {
        config_content = fs::read_to_string(".auto-commit-msg.toml")?;
    }
    let config: Config = toml::from_str(&config_content)?;

    let output = Command::new("git").args(["diff", "--cached"]).output()?;
    let diff = String::from_utf8(output.stdout)?;

    let provider = config.provider;
    let base_url = reqwest::Url::parse(&provider.base_url)?;
    let token = std::env::var(&provider.api_key)?;
    let client = OpenAIClient::build(base_url, token)?;

    let stat_output = Command::new("git")
        .args(["diff", "--cached", "--shortstat"])
        .output()?;
    let stat = String::from_utf8(stat_output.stdout)?;
    let re = Regex::new(
        r"
        (?:,\s+(?P<insertions>\d+)\s+insertions?\(\+\))?
        (?:,\s+(?P<deletions>\d+)\s+deletions?\(-\))?
    ",
    )?;

    let mut insertions: u32 = 0;
    let mut deletions: u32 = 0;
    if let Some(caps) = re.captures(&stat) {
        insertions = caps.name("insertions").unwrap().as_str().parse()?;
        deletions = caps.name("deletions").unwrap().as_str().parse()?;
    }
    let total_changes = insertions + deletions;

    let diff_config = config.diff;
    let mut model = diff_config.short_model;
    if total_changes >= diff_config.threshold {
        model = diff_config.long_model;
        info!("Using long model {model}")
    }

    let mut response_duration = None;
    if config.trace {
        response_duration = Some(Instant::now());
    }
    let completion = client.create_chat_completion(Chat {
        model: model.clone(),
        messages: vec![
            ChatMessage {
                role: "developer".to_string(),
                content: r#"
				    You are an assistant that writes concise, conventional commit
                    messages based on the provided git diff. Return the commit
                    message without any quotes.
                    "#
                .to_string(),
            },
            ChatMessage {
                role: "user".to_string(),
                content: diff,
            },
        ],
    })?;
    let response_time = response_duration.map(|foo| foo.elapsed());

    let messages: Vec<&ChoiceMessage> = completion
        .choices
        .iter()
        .filter(|choice| choice.message.role == "assistant")
        .map(|choice| &choice.message)
        .collect();
    let message = messages.first().expect("at least one message is expected");
    let mut commit_msg = message.content.clone();
    if let Some(bar) = response_time {
        commit_msg.push_str("\n---\n");
        commit_msg.push_str(&serde_json::to_string(&TraceWrapper(Trace {
            language: "rust".to_string(),
            // TODO: avoid using clone
            model: model.clone(),
            response_time: TraceDuration(bar),
            execution_time: TraceDuration(execution_duration.elapsed()),
        }))?);
    }

    if let Some(commit_msg_file) = env::args().nth(1) {
        fs::write(commit_msg_file, commit_msg)?;
    } else {
        println!("{commit_msg}");
    }

    Ok(())
}
