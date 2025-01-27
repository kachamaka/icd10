import sys
import nltk
from nltk.corpus import stopwords
from nltk.tokenize import word_tokenize
from nltk.stem import WordNetLemmatizer
import string
import re

# Initialize the lemmatizer
lemmatizer = WordNetLemmatizer()

# Download NLTK resources
nltk.download('punkt')
nltk.download('punkt_tab')
nltk.download('stopwords')
nltk.download('wordnet')

def process_text(text):
    # Tokenize the text into words
    words = word_tokenize(text.lower())  # Convert text to lowercase first for consistency

    # Remove contractions and unwanted parts (like 've', 'm', etc.)
    contractions = ["'ve", "'m", "'ll", "'", "’", "'s", "n't", "'re"]
    words = [word for word in words if word not in contractions]
    
    # Remove stop words and punctuation
    stop_words = set(stopwords.words('english'))
    punctuation = set(string.punctuation)
    
    filtered_words = [word for word in words if word not in stop_words and word not in punctuation]
    
    # Lemmatize the words
    lemmatized_words = [lemmatizer.lemmatize(word) for word in filtered_words]
    
    return lemmatized_words


if __name__ == "__main__":
    # Example input text (can also be passed from command line)
    # text = "I've been feeling really tired for the past few days, and today I woke up with a headache and a sore throat."
    text = "Yesterday I had mild headache. Today I have a high temperature and a strong stomachache. And now I have a headache and a stomachache."
    # text = "I woke up with a mild headache and a sore throat. Throughout the day, I experienced a mild fever and some dizziness. In the evening, my body temperature was higher than usual, and I also had a stomachache. The headache lasted all day, and now I feel exhausted."
    # text = "Today, I noticed a sharp pain in my lower back and a stiff neck. I also feel some fatigue and nausea, and I have a slight fever. The pain in my back seems to worsen when I move, and the nausea makes it hard to eat."
    # text = "After a long day at work, I developed a headache and a tight chest. I had a low-grade fever and some difficulty breathing, especially when I climbed stairs. I feel lightheaded and weak, and my joints are aching."
    # text = "I've been feeling really tired for the past few days, and today I woke up with a headache and a sore throat. Yesterday, I had a mild fever that lasted for a few hours, but it went away in the evening. I also have a persistent cough and some difficulty breathing, especially at night."
    # text = "Lately, I’ve been feeling nauseous and have had some stomach cramps. My head is throbbing with pain, and I’m also experiencing shortness of breath and dizziness. I feel exhausted and have trouble sleeping at night."
    # text = "This morning, I had a sharp stomachache and some bloating. The pain is localized around my stomach, and I've also been feeling nauseous. I didn’t eat much last night, and now I have a headache that feels like pressure behind my eyes."
    # text = "I feel a bit weak today, and my throat is scratchy. I also have a headache that’s making it hard to concentrate. My body temperature feels slightly elevated, and I noticed some mild swelling in my legs."
    input_text = sys.argv[1] if len(sys.argv) > 1 else text

    # Process the input text
    processed_words = process_text(input_text)
    
    # Join the processed words with a space and output the result
    print(" ".join(processed_words))

