package pro290.vaporgame.PRO290VaporGameAPI.Controller;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import pro290.vaporgame.PRO290VaporGameAPI.Modle.Game;

import java.util.ArrayList;

@RestController
@RequestMapping(value = "/game")
public class GameRestController {

    @GetMapping("/")
    public Game GetExampleGame() {
        return new Game("League Of Legends",
                "A 2009 multiplayer online battle arena video game developed and published by Riot Games. " +
                            "Inspired by Defense of the Ancients, a custom map for Warcraft III, Riot's founders sought " +
                            "to develop a stand-alone game in the same genre.",
                    "Riot Games");
    }
}
