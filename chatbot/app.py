#!/usr/bin/env python3
import os
import json
import logging
from flask import Flask, request, jsonify, make_response, render_template
import openai
import llm

# Constants for chat message roles and model name
CHAT_MESSAGE_ROLE_SYSTEM = "system"
CHAT_MESSAGE_ROLE_USER = "user"
CHAT_MESSAGE_ROLE_ASSISTANT = "assistant"
GPT4oMini = "gpt-4o-mini"

# Global chatbot variable
chatbot = None

# Ensure the OpenAI API key is set
openai.api_key = os.getenv('OPENAI_PROJECT_KEY')

# Function returns the tool definition for web search
def WebSearchTool():
    return {
        "name": "web_search",
        "description": "Performs a web search for the query",
        "parameters": {
            "type": "object",
            "properties": {
                "query": {
                    "type": "string",
                    "description": "Search query"
                }
            },
            "required": ["query"]
        }
    }

# ChatCompletionClient provides the CreateChatCompletion method
class ChatCompletionClient:
    def CreateChatCompletion(self, req):
        try:
            response = openai.ChatCompletion.create(
                model=req["Model"],
                messages=req["Messages"]
            )
            
            return response, None
        except Exception as e:
            logging.error("Error in CreateChatCompletion: %s", e)
            return None, e

# LLMClient wraps the OpenAI API client
class LLMClient:
    def __init__(self):
        self.client = ChatCompletionClient()  # This is correctly initialized

    def CreateChatCompletion(self, req):  
        return self.client.CreateChatCompletion(req)  


# ChatBot handles user interactions and uses LLMClient to fetch responses
class ChatBot:
    def __init__(self, llm_client, system_msg=""):
        self.llm_client = llm_client
        self.system_msg = system_msg or "You are a helpful assistant for general queries."
        self.tools = [WebSearchTool()]
        self.conversation_history = [
            {"role": CHAT_MESSAGE_ROLE_SYSTEM, "content": self.system_msg}
        ]

    def AnswerQuestion(self, question):
        self.conversation_history.append({"role": CHAT_MESSAGE_ROLE_USER, "content": question})
        req = {
            "Model": GPT4oMini,
            "Messages": self.conversation_history
        }
        response, err = self.llm_client.CreateChatCompletion(req)
        if err:
            return "", err
        
        llm_response = response.choices[0].message["content"] if response.choices else "I couldn't find an answer to your question."
        return llm_response, None

def InitializeChatBot():
    llm_client = LLMClient()
    global chatbot
    chatbot = ChatBot(llm_client)

app = Flask(__name__)

@app.route('/')
def home():
    return render_template('index.html')

@app.route('/chatbot', methods=['GET', 'POST'])
def chatbot_route():
    print("Accessed /chatbot route with method:", request.method)  # Log the request method
    if request.method == 'POST':
        data = request.get_json() or {}
        print("Received POST data:", data)  # Log the data received in POST request
        
        question = data.get('question')
        if not question:
            print("Error: No question provided in POST data")  # Log missing question error
            return make_response("Question is required.", 400)
        
        response, err = chatbot.AnswerQuestion(question)
        if err:
            print("Error during chatbot response generation:", err)  # Log error from chatbot
            return make_response(str(err), 500)
        
        print("Chatbot response:", response)  # Log successful chatbot response
        return jsonify({"response": response})

    else:
        print("GET request received, serving index.html")  # Log GET request handling
        return render_template('index.html')


if __name__ == '__main__':
    logging.basicConfig(level=logging.INFO)
    InitializeChatBot()
    app.run(host='0.0.0.0', port=5001, debug=True)
