using AutoMapper;
public class AutoMapperProfile : Profile
{
    public AutoMapperProfile()
    {
        // User mappings
        CreateMap<UserDTO, User>().ReverseMap();
        
        // Game mappings
        CreateMap<GameDTO, Game>().ReverseMap();
        
        // Order mappings
        CreateMap<OrderDTO, Order>()
            .ForMember(dest => dest.OrderGames, opt => opt.MapFrom(src => src.Games.Select(g => new OrderGame { GameID = g.ID })))
            .ReverseMap()
            .ForMember(dest => dest.Games, opt => opt.MapFrom(src => src.OrderGames.Select(og => og.Game)));
  
    }
}
