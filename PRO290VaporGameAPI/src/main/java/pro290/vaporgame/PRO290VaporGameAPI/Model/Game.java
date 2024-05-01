package pro290.vaporgame.PRO290VaporGameAPI.Model;

import java.time.LocalDate;
import java.util.ArrayList;
import java.util.UUID;

public class Game {

    UUID Id;
    String Title;
    String Description;
    String Author;
    ArrayList<String> Tags;
    String CreationDate;

    public Game(String title, String description, String author, ArrayList<String> tags) {
        this.Id = UUID.randomUUID();
        Title = title;
        Description = description;
        Author = author;
        Tags = tags;
        CreationDate = LocalDate.now().toString();
    }

    public Game(String title, String description, String author) {
        this.Id = UUID.randomUUID();
        Title = title;
        Description = description;
        Author = author;
        LocalDate.now().toString();
    }
}
