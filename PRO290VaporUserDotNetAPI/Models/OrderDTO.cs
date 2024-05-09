using System;
using System.ComponentModel.DataAnnotations;
 
public class OrderDTO
{
    public Guid CartGuid { get; set; }
    public Boolean ReadCart { get; set; }
    public List<GameDTO>? Games { get; set; }
}
