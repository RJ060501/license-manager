import os

from dotenv import load_dotenv

def load_settings():
    load_dotenv()

    settings = {
        "tenant_id": os.getenv("TENANT_ID"),
        "client_id": os.getenv("CLIENT_ID"),
        "client_secret": os.getenv("CLIENT_SECRET")
    }
    
    missing = [
        key
        for key, value in settings.items()
        if not value
    ]

    if missing:
        raise ValueError(
            f"Missing required environment variables: {', '.join(missing)}"
        )

    return settings