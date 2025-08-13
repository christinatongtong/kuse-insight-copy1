from pinecone import Pinecone


indexHost = "https://documents-1ucbdkk.svc.gcp-us-central1-4a9f.pinecone.io"
apiKey = "00844b4c-14fc-4ab5-a700-cb90ad18b20e"
namespace = "documents"

pc = Pinecone(api_key=apiKey)

# To get the unique host for an index,
# see https://docs.pinecone.io/guides/manage-data/target-an-index
index = pc.Index(host=indexHost)

filtered_results = index.search(
    namespace=namespace,
    query={
        "inputs": {"text": ""},
        "top_k": 3,
        "filter": {"user_id": 108},
    },
    fields=["user_id"],
)

print(filtered_results)
