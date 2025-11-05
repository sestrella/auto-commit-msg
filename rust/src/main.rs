use anyhow::Result;
use reqwest::header;
use serde::ser::SerializeStruct;
use std::process::Command;
use std::time::{Duration, Instant};
use std::{env, fs};

struct OpenAIClient {
    client: reqwest::blocking::Client,
    base_url: reqwest::Url,
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

struct Trace {
    model: String,
    response_time: TraceDuration,
    execution_time: TraceDuration,
}

impl serde::Serialize for Trace {
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let mut state = serializer.serialize_struct("Trace", 3)?;
        state.serialize_field("model", &self.model)?;
        state.serialize_field("response_time", &self.response_time)?;
        state.serialize_field("execution_time", &self.execution_time)?;
        state.end()
    }
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

    let output = Command::new("git").arg("diff").arg("--cached").output()?;
    let diff = String::from_utf8(output.stdout)?;

    let base_url = reqwest::Url::parse("https://generativelanguage.googleapis.com/v1beta/openai/")?;
    let token = std::env::var("GEMINI_API_KEY")?;
    let client = OpenAIClient::build(base_url, token)?;

    let response_duration = Instant::now();
    let completion = client.create_chat_completion(Chat {
        model: "gemini-2.5-flash-lite".to_string(),
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
    let response_time = response_duration.elapsed();

    let messages: Vec<&ChoiceMessage> = completion
        .choices
        .iter()
        .filter(|choice| choice.message.role == "assistant")
        .map(|choice| &choice.message)
        .collect();
    let message = messages.first().expect("TODO");
    let mut commit_msg = message.content.clone();
    commit_msg.push_str("\n---\n\"auto-commit-msg\":");
    commit_msg.push_str(&serde_json::to_string(&Trace {
        model: "gemini-2.5-flash-lite".to_string(),
        response_time: TraceDuration(response_time),
        execution_time: TraceDuration(execution_duration.elapsed()),
    })?);

    if let Some(commit_msg_file) = env::args().nth(1) {
        fs::write(commit_msg_file, commit_msg)?;
    } else {
        println!("{commit_msg}");
    }

    Ok(())
}
