import base64
import os 
import requests
import json

def get_cluster_list():
     """
     get_cluster_list will call CastAI and get the list of clusters in order to match 
          against the cluster_name passed in the event payload, this prevents the user 
          from needing to know the cluster_id's of the clusters to be actioned. 
     """
     url = "https://api.cast.ai/v1/kubernetes/external-clusters"
     headers={"Content-Type": 'application/json', "X-API-Key": os.environ['CastAIKey'] }
     response = requests.get(url, headers=headers)
     if response.status_code == 200:
          print("Successfully retrieved cluster list")
          return response.json()
     else: 
          print("Failed to receive cluster list, unable to proceed, status code: %d, response text %s"%(response.status_code, response.text))
          return None

def action_cluster(cluster_id, action):
     """ 
     action_cluster will call the CastAI API with the cluster_id to be 
          actioned upon and the appropriate action (pause/resume)
          The result will be logged for analysis if needed. 
     """
     url = "https://api.cast.ai/v1/kubernetes/external-clusters/" + cluster_id + "/" + action
     headers={"Content-Type": 'application/json', "X-API-Key": os.environ['CastAIKey'] }
     response = requests.post(url, headers=headers)
     if response.status_code == 200:
          print("Successfully %sd cluster ID: %s"%(action, cluster_id))
     else: 
          print("Failed to %s the cluster ID: %s, status code: %d"%(action, cluster_id, response.status_code))

def cluster_action_handler(event, context):
    """Triggered from a message on a Cloud Pub/Sub topic.
    Args:
         event (dict): Event payload.
         context (google.cloud.functions.Context): Metadata for the event.
    """
    action_payload = json.loads(base64.b64decode(event['data']).decode('utf-8'))
    
    cluster_list = get_cluster_list()
    if cluster_list == None:
         return None
    clusters_to_action = []
    for cl in cluster_list["items"]: 
         for cluster_name in action_payload["clusterNames"]:
              if cl["name"] == cluster_name:
                   action_cluster(cl["id"], action_payload["action"])