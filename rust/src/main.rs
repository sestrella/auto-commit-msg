use anyhow::Result;
use reqwest::header;
use std::process::Command;

struct OpenAIClient {
    client: reqwest::Client,
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
    message: Message,
}

#[derive(serde::Deserialize)]
struct Message {
    content: String,
}

impl OpenAIClient {
    fn build(base_url: reqwest::Url, token: String) -> Result<Self> {
        let mut headers = header::HeaderMap::new();
        headers.insert(header::AUTHORIZATION, format!("Bearer {token}").parse()?);
        let client = reqwest::Client::builder()
            .default_headers(headers)
            .build()?;
        Ok(Self { client, base_url })
    }

    async fn create_chat_completion(&self, chat: Chat) -> Result<Completion> {
        let url = self.base_url.join("chat/completions")?;
        let completion = self
            .client
            .post(url)
            .json(&chat)
            .send()
            .await?
            .json()
            .await?;
        Ok(completion)
    }
}

#[tokio::main]
async fn main() -> Result<()> {
    let output = Command::new("git").arg("diff").arg("--cached").output()?;
    let diff = String::from_utf8(output.stdout)?;

    let base_url = reqwest::Url::parse("https://generativelanguage.googleapis.com/v1beta/openai/")?;
    let token = std::env::var("GEMINI_API_KEY")?;
    let client = OpenAIClient::build(base_url, token)?;
    let resp = client
        .create_chat_completion(Chat {
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
        })
        .await?;
    println!("{}", resp.choices[0].message.content);

    Ok(())
}
