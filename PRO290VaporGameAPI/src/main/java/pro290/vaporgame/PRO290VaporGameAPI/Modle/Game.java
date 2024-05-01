package pro290.vaporgame.PRO290VaporGameAPI.Modle;

import java.time.LocalDate;
import java.util.ArrayList;
import java.util.UUID;

public class Game {

    UUID Id = UUID.randomUUID();
    String Title;
    String Description;
    String Author;
    ArrayList<String> Tags;
    String CreationDate = LocalDate.now().toString();

    public Game(String title, String description, String author, ArrayList<String> tags) {
        Title = title;
        Description = description;
        Author = author;
        Tags = tags;
    }

    public Game(String title, String description, String author) {
        Title = title;
        Description = description;
        Author = author;
    }
}
