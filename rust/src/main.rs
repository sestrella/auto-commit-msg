use anyhow::Result;
use reqwest::header;

struct OpenAIClient {
    client: reqwest::Client,
    base_url: reqwest::Url,
}

#[derive(Debug, serde::Deserialize)]
struct Completion {}

impl OpenAIClient {
    fn build(token: String, base_url: reqwest::Url) -> Result<Self> {
        let mut headers = header::HeaderMap::new();
        headers.insert(header::AUTHORIZATION, format!("Bearer {}", token).parse()?);
        let client = reqwest::Client::builder()
            
            
            .default_headers(headers).build()?;
        Ok(Self { client, base_url })
    }

    async fn create_chat_completion(&self) -> Result<Completion> {
        let url = self.base_url.join("/chat/completions")?;
        let completion = self.client.post(url).send().await?.json().await?;
        Ok(completion)
    }
}

#[tokio::main]
async fn main() -> Result<()> {
    let token = std::env::var("GEMINI_API_KEY")?;
    let base_url = reqwest::Url::parse("https://generativelanguage.googleapis.com/v1beta/openai")?;
    let client = OpenAIClient::build(token, base_url)?;
    let resp = client.create_chat_completion().await?;
    println!("{resp:#?}");
    Ok(())
}
