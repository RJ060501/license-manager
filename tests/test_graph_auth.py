import os

from azure.identity import ClientSecretCredential
from dotenv import load_dotenv


load_dotenv()


tenant_id = os.getenv("TENANT_ID")
client_id = os.getenv("CLIENT_ID")
client_secret = os.getenv("CLIENT_SECRET")


if not tenant_id or not client_id or not client_secret:
    raise ValueError(
        "Missing TENANT_ID, CLIENT_ID, or CLIENT_SECRET in the .env file."
    )


credential = ClientSecretCredential(
    tenant_id=tenant_id,
    client_id=client_id,
    client_secret=client_secret,
)


token = credential.get_token("https://graph.microsoft.com/.default")


print("Authentication successful.")
print(f"Token expires at: {token.expires_on}")