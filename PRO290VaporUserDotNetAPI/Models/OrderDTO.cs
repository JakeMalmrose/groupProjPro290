using System;
using System.Collections.Generic;

public class OrderDTO
{
    public Guid CartGuid { get; set; }

    public bool ReadCart { get; set; }

    public List<GameDTO>? Games { get; set; }
}
