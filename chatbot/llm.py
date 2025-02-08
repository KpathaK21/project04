import openai
import logging
import re

class LLMClient:
    """Wraps the OpenAI client for interaction with an LLM."""
    
    def __init__(self, api_key: str):
        """Initializes a new LLMClient with the provided OpenAI API key."""
        self.client = openai.OpenAI(api_key=api_key)

    def chat_completion(self, question: str, system_message: str) -> str:
        """Sends a user's query to the LLM and retrieves a response."""
        try:
            response = self.client.chat.completions.create(
                model="gpt-4",  # Use GPT-4 model
                messages=[
                    {"role": "system", "content": system_message},
                    {"role": "user", "content": question},
                ]
            )
            return response.choices[0].message.content
        except Exception as e:
            raise RuntimeError(f"ChatCompletion failed: {e}")

class ChatCompletionClient:
    """Handles chat completion requests using the OpenAI API."""
    
    def chat_completion(self, question: str, system_message: str) -> str:
        """Sends a user's query to the LLM and retrieves a response with enhanced formatting."""
        try:
            response = self.client.chat.completions.create(
                model="gpt-4",  # Use GPT-4 model
                messages=[
                    {"role": "system", "content": system_message},
                    {"role": "user", "content": question},
                ]
            )
    
            # Extract the raw response
            raw_response = response.choices[0].message.content
    
            # Ensure the response doesn't start with the echoed question
            if raw_response.startswith(question):
                raw_response = raw_response[len(question):].strip()
    
            # Remove unnecessary bot prefixes
            cleaned_response = raw_response.replace("Bot:", "").strip()
    
            # Apply formatting rules
            formatted_response = format_response(cleaned_response)
    
            return formatted_response
    
        except Exception as e:
            raise RuntimeError(f"ChatCompletion failed: {e}")
    
    
    def format_response(text: str) -> str:
            """
            Formats response to maintain proper Markdown structure and readability.
            Preserves code blocks, headers, and list formatting while ensuring proper spacing.
            """
            # Split the text into lines
            lines = text.split('\n')
            formatted_lines = []
            in_code_block = False
            
            for i, line in enumerate(lines):
                # Handle code blocks
                if line.startswith('```'):
                    in_code_block = not in_code_block
                    formatted_lines.append(line)
                    if not in_code_block:  # End of code block
                        formatted_lines.append('')
                    continue
                    
                if in_code_block:
                    formatted_lines.append(line)
                    continue
                    
                # Handle headers
                if line.startswith('#'):
                    if i > 0:  # Add blank line before header if not first line
                        formatted_lines.append('')
                    formatted_lines.append(line)
                    formatted_lines.append('')
                    continue
                    
                # Handle list items
                if line.strip().startswith(('-', '*', '1.')):
                    formatted_lines.append(line)
                    continue
                    
                # Handle regular text
                if line.strip():
                    formatted_lines.append(line)
                elif i > 0 and formatted_lines and formatted_lines[-1]:
                    formatted_lines.append('')
            
            # Join lines and clean up multiple blank lines
            text = '\n'.join(formatted_lines)
            text = re.sub(r'\n{3,}', '\n\n', text)
            
            return text.strip()
    
